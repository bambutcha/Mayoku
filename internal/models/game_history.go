package models

import (
	"time"
)

// GameHistory представляет историю завершенной игры
type GameHistory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RoomUUID  string    `gorm:"index;not null" json:"room_uuid"` // UUID комнаты для связи с логами
	DeckName  string    `gorm:"not null" json:"deck_name"`       // Название колоды
	Winner    string    `gorm:"not null" json:"winner"`          // "Spy" | "Locals"
	Duration  int       `gorm:"not null" json:"duration"`        // Длительность в секундах
	CreatedAt time.Time `json:"created_at"`

	// Связи
	Players []User `gorm:"many2many:game_players;" json:"players,omitempty"`
}

// TableName задает имя таблицы
func (GameHistory) TableName() string {
	return "game_history"
}
