package models

import (
	"time"
)

// DeckStatus представляет статус модерации колоды
type DeckStatus string

const (
	DeckStatusDraft    DeckStatus = "draft"    // Черновик
	DeckStatusPending  DeckStatus = "pending"  // На модерации
	DeckStatusApproved DeckStatus = "approved" // Одобрено
	DeckStatusRejected DeckStatus = "rejected" // Отклонено
)

// Deck представляет набор локаций (колоду)
type Deck struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	AuthorID  uint       `gorm:"not null;index" json:"author_id"`
	Name      string     `gorm:"not null" json:"name"`
	IsPublic  bool       `gorm:"default:false" json:"is_public"`
	Status    DeckStatus `gorm:"type:varchar(20);default:'draft'" json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	// Связи
	Author    User       `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Locations []Location `gorm:"foreignKey:DeckID;constraint:OnDelete:CASCADE" json:"locations,omitempty"`
}

// TableName задает имя таблицы
func (Deck) TableName() string {
	return "decks"
}
