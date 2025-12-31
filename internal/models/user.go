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

	// Statistics
	GamesPlayed int `gorm:"default:0" json:"games_played"`
	WinsSpy     int `gorm:"default:0" json:"wins_spy"`
	WinsLocal   int `gorm:"default:0" json:"wins_local"`
	LossesSpy   int `gorm:"default:0" json:"losses_spy"`
	LossesLocal int `gorm:"default:0" json:"losses_local"`

	// Связи
	Decks       []Deck        `gorm:"foreignKey:AuthorID" json:"decks,omitempty"`
	GameHistory []GameHistory `gorm:"many2many:game_players;" json:"game_history,omitempty"`
}

// TableName задает имя таблицы
func (User) TableName() string {
	return "users"
}
