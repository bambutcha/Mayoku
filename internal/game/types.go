package game

import (
	"encoding/json"
	"time"
)

// PlayerRole представляет роль игрока
type PlayerRole string

const (
	RoleSpy   PlayerRole = "spy"
	RoleLocal PlayerRole = "local"
)

// GameStatus представляет статус игры
type GameStatus string

const (
	StatusWaiting  GameStatus = "waiting"  // Ожидание игроков
	StatusPlaying  GameStatus = "playing"  // Игра идет
	StatusVoting   GameStatus = "voting"   // Идет голосование
	StatusFinished GameStatus = "finished" // Игра завершена
)

// Player представляет игрока в комнате
type Player struct {
	UserID       uint       `json:"user_id"`
	TgID         int64      `json:"tg_id"`
	Username     string     `json:"username"`
	AvatarURL    string     `json:"avatar_url"`
	Role         PlayerRole `json:"role,omitempty"`          // Скрыто до начала игры
	Location     string     `json:"location,omitempty"`      // Только для Local
	LocationRole string     `json:"location_role,omitempty"` // Роль в локации (только для Local)
	IsReady      bool       `json:"is_ready"`
	IsVoted      bool       `json:"is_voted,omitempty"` // Для голосования
	Vote         bool       `json:"vote,omitempty"`     // true = за, false = против
}

// LocationInfo информация о локации
type LocationInfo struct {
	Name     string   `json:"name"`
	ImageURL string   `json:"image_url"`
	Roles    []string `json:"roles"`
}

// VotingState состояние голосования
type VotingState struct {
	TargetUserID uint          `json:"target_user_id"`
	Votes        map[uint]bool `json:"votes"` // user_id -> vote (true/false)
	StartedAt    time.Time     `json:"started_at"`
}

// RoomState состояние комнаты
type RoomState struct {
	RoomID     string           `json:"room_id"`
	Status     GameStatus       `json:"status"`
	Players    map[uint]*Player `json:"players"` // user_id -> Player
	Location   *LocationInfo    `json:"location,omitempty"`
	SpyIDs     []uint           `json:"spy_ids,omitempty"` // ID шпионов
	TimerEnd   *time.Time       `json:"timer_end,omitempty"`
	Voting     *VotingState     `json:"voting,omitempty"`
	Winner     string           `json:"winner,omitempty"` // "spy" | "locals"
	DeckID     uint             `json:"deck_id"`
	DeckName   string           `json:"deck_name"`
	MaxPlayers int              `json:"max_players"`
	SpyCount   int              `json:"spy_count"` // Количество шпионов
	Duration   int              `json:"duration"`  // В минутах
	CreatedBy  uint             `json:"created_by"`
	CreatedAt  time.Time        `json:"created_at"`
}

// WSMessage сообщение WebSocket
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// ClientMessage сообщение от клиента
type ClientMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
