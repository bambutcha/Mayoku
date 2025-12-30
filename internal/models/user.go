package models

import (
	"time"
)

// User представляет пользователя Telegram
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TgID      int64     `gorm:"uniqueIndex;not null" json:"tg_id"`
	Username  string    `gorm:"index" json:"username"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Связи
	Decks []Deck `gorm:"foreignKey:AuthorID" json:"decks,omitempty"`
}

// TableName задает имя таблицы
func (User) TableName() string {
	return "users"
}
