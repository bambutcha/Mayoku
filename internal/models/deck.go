package models

import (
	"time"
)

// Deck представляет набор локаций (колоду)
type Deck struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	AuthorID  uint      `gorm:"not null;index" json:"author_id"`
	Name      string    `gorm:"not null" json:"name"`
	IsPublic  bool      `gorm:"default:false" json:"is_public"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Связи
	Author    User       `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Locations []Location `gorm:"foreignKey:DeckID;constraint:OnDelete:CASCADE" json:"locations,omitempty"`
}

// TableName задает имя таблицы
func (Deck) TableName() string {
	return "decks"
}
