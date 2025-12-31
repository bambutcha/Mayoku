package game

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	logger "github.com/Chelaran/yagalog"
)

var (
	ErrRoomExists = &GameError{Message: "room already exists"}
	ErrRoomNotFound = &GameError{Message: "room not found"}
	ErrPlayerNotFound = &GameError{Message: "player not found"}
)

type GameError struct {
	Message string
}

func (e *GameError) Error() string {
	return e.Message
}

// Client представляет WebSocket клиента
type Client struct {
	mu       sync.Mutex
	conn     *websocket.Conn
	send     chan []byte
	hub      *Hub
	room     *Room
	userID   uint
	tgID     int64
	username string
	avatarURL string
	log      *logger.Logger
}

// NewClient создает нового клиента
func NewClient(conn *websocket.Conn, hub *Hub, userID uint, tgID int64, username, avatarURL string) *Client {
	log, _ := logger.NewLogger()
	return &Client{
		conn:      conn,
		send:      make(chan []byte, 256),
		hub:       hub,
		userID:    userID,
		tgID:      tgID,
		username:  username,
		avatarURL: avatarURL,
		log:       log,
	}
}

// ReadPump читает сообщения из WebSocket
func (c *Client) ReadPump() {
	defer func() {
		c.conn.Close()
		if c.room != nil {
			c.room.RemovePlayer(c.userID)
		}
	}()

	for {
		var msg ClientMessage
		if err := c.conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.log.Error("WebSocket error: %v", err)
			}
			break
		}

		c.handleMessage(msg)
	}
}

// WritePump отправляет сообщения в WebSocket
func (c *Client) WritePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.log.Error("Failed to write message: %v", err)
				return
			}
		}
	}
}

// SendMessage отправляет сообщение клиенту
func (c *Client) SendMessage(msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		c.log.Error("Failed to marshal message: %v", err)
		return
	}

	c.SendRaw(data)
}

// SendRaw отправляет сырые данные
func (c *Client) SendRaw(data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case c.send <- data:
	default:
		c.log.Warning("Client send buffer full, dropping message")
	}
}

// SendError отправляет ошибку клиенту
func (c *Client) SendError(err error) {
	msg := WSMessage{
		Type: "error",
		Payload: map[string]string{
			"message": err.Error(),
		},
	}
	c.SendMessage(msg)
}

// handleMessage обрабатывает входящее сообщение
func (c *Client) handleMessage(msg ClientMessage) {
	switch msg.Type {
	case "join_room":
		c.handleJoinRoom(msg.Payload)
	case "set_ready":
		c.handleSetReady(msg.Payload)
	case "vote_start":
		c.handleVoteStart(msg.Payload)
	case "vote_answer":
		c.handleVoteAnswer(msg.Payload)
	case "spy_guess":
		c.handleSpyGuess(msg.Payload)
	case "kick_player":
		c.handleKickPlayer(msg.Payload)
	default:
		c.SendError(&GameError{Message: "unknown message type"})
	}
}

// handleKickPlayer обрабатывает исключение игрока из комнаты (только админ комнаты)
func (c *Client) handleKickPlayer(payload json.RawMessage) {
	if c.room == nil {
		c.SendError(&GameError{Message: "not in a room"})
		return
	}

	var req struct {
		TargetUserID uint `json:"target_user_id"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		c.SendError(&GameError{Message: "invalid payload"})
		return
	}

	if err := c.room.KickPlayer(c.userID, req.TargetUserID); err != nil {
		c.SendError(err)
		return
	}
}

// handleJoinRoom обрабатывает присоединение к комнате
func (c *Client) handleJoinRoom(payload json.RawMessage) {
	var req struct {
		RoomID string `json:"room_id"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		c.SendError(&GameError{Message: "invalid payload"})
		return
	}

	room, exists := c.hub.GetRoom(req.RoomID)
	if !exists {
		c.SendError(ErrRoomNotFound)
		return
	}

	// Добавляем игрока в комнату
	if err := room.AddPlayer(c.userID, c.tgID, c.username, c.avatarURL, c); err != nil {
		c.SendError(err)
		return
	}

	c.room = room

	// Отправляем подтверждение с информацией о правах
	isRoomAdmin := room.IsRoomAdmin(c.userID)
	c.SendMessage(WSMessage{
		Type: "joined_room",
		Payload: map[string]interface{}{
			"room_id":      req.RoomID,
			"is_room_admin": isRoomAdmin,
		},
	})
}

// handleSetReady обрабатывает установку готовности
func (c *Client) handleSetReady(payload json.RawMessage) {
	if c.room == nil {
		c.SendError(&GameError{Message: "not in a room"})
		return
	}

	var req struct {
		Ready bool `json:"ready"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		c.SendError(&GameError{Message: "invalid payload"})
		return
	}

	if err := c.room.SetPlayerReady(c.userID, req.Ready); err != nil {
		c.SendError(err)
		return
	}
}

// handleVoteStart обрабатывает начало голосования
func (c *Client) handleVoteStart(payload json.RawMessage) {
	if c.room == nil {
		c.SendError(&GameError{Message: "not in a room"})
		return
	}

	var req struct {
		TargetUserID uint `json:"target_user_id"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		c.SendError(&GameError{Message: "invalid payload"})
		return
	}

	if err := c.room.StartVoting(c.userID, req.TargetUserID); err != nil {
		c.SendError(err)
		return
	}
}

// handleVoteAnswer обрабатывает ответ на голосование
func (c *Client) handleVoteAnswer(payload json.RawMessage) {
	if c.room == nil {
		c.SendError(&GameError{Message: "not in a room"})
		return
	}

	var req struct {
		Vote bool `json:"vote"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		c.SendError(&GameError{Message: "invalid payload"})
		return
	}

	if err := c.room.Vote(c.userID, req.Vote); err != nil {
		c.SendError(err)
		return
	}
}

// handleSpyGuess обрабатывает попытку шпиона угадать локацию
func (c *Client) handleSpyGuess(payload json.RawMessage) {
	if c.room == nil {
		c.SendError(&GameError{Message: "not in a room"})
		return
	}

	var req struct {
		LocationName string `json:"location_name"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		c.SendError(&GameError{Message: "invalid payload"})
		return
	}

	if err := c.room.SpyGuess(c.userID, req.LocationName); err != nil {
		c.SendError(err)
		return
	}
}

