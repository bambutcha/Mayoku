package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// StringArray представляет массив строк для JSON сериализации в GORM
type StringArray []string

// Value реализует driver.Valuer для сохранения в БД
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	return json.Marshal(a)
}

// Scan реализует sql.Scanner для чтения из БД
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal StringArray value")
	}

	return json.Unmarshal(bytes, a)
}

// Location представляет локацию в игре
type Location struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	DeckID    uint        `gorm:"not null;index" json:"deck_id"`
	Name      string      `gorm:"not null" json:"name"`
	ImageURL  string      `json:"image_url"`               // Ссылка на MinIO
	Roles     StringArray `gorm:"type:jsonb" json:"roles"` // ["Доктор", "Медсестра", ...]
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`

	// Связи
	Deck Deck `gorm:"foreignKey:DeckID" json:"deck,omitempty"`
}

// TableName задает имя таблицы
func (Location) TableName() string {
	return "locations"
}
