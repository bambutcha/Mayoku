package game

import (
	"sync"

	logger "github.com/Chelaran/yagalog"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Hub управляет всеми комнатами и клиентами
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]*Room // room_id -> Room
	db    *gorm.DB
	redis *redis.Client
	log   *logger.Logger
}

// NewHub создает новый Hub
func NewHub(db *gorm.DB, redis *redis.Client) *Hub {
	log, _ := logger.NewLogger()
	return &Hub{
		rooms: make(map[string]*Room),
		db:    db,
		redis: redis,
		log:   log,
	}
}

// CreateRoom создает новую комнату
func (h *Hub) CreateRoom(roomID string, createdBy uint, deckID uint, deckName string, maxPlayers, spyCount, duration int) (*Room, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Проверяем, не существует ли уже
	if _, exists := h.rooms[roomID]; exists {
		return nil, ErrRoomExists
	}

	room := NewRoom(roomID, createdBy, deckID, deckName, maxPlayers, spyCount, duration, h.db, h.redis)
	h.rooms[roomID] = room

	h.log.Info("Room created: %s (created by: %d)", roomID, createdBy)

	return room, nil
}

// GetRoom возвращает комнату по ID
func (h *Hub) GetRoom(roomID string) (*Room, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, exists := h.rooms[roomID]
	return room, exists
}

// DeleteRoom удаляет комнату
func (h *Hub) DeleteRoom(roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.rooms, roomID)
	h.log.Info("Room deleted: %s", roomID)
}

// ListRooms возвращает список активных комнат
func (h *Hub) ListRooms() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	rooms := make([]string, 0, len(h.rooms))
	for id := range h.rooms {
		rooms = append(rooms, id)
	}

	return rooms
}
