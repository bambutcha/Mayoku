package utils

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	initdata "github.com/telegram-mini-apps/init-data-golang"
)

// VerifyTelegramInitData проверяет подпись Telegram initData
// initData - строка вида "query_id=...&user=...&auth_date=...&hash=..."
// botToken - токен бота от BotFather
func VerifyTelegramInitData(initData, botToken string) (bool, error) {
	if botToken == "" {
		return false, fmt.Errorf("bot token is empty")
	}

	// Используем готовую библиотеку для проверки
	// TTL = 24 часа (86400 секунд)
	err := initdata.Validate(initData, botToken, 24*time.Hour)
	if err != nil {
		return false, fmt.Errorf("validation failed: %w", err)
	}

	return true, nil
}

// ParseTelegramUser извлекает данные пользователя из initData
type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty"`
}

func ParseTelegramUser(initData string) (*TelegramUser, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse initData: %w", err)
	}

	userStr := values.Get("user")
	if userStr == "" {
		return nil, fmt.Errorf("user not found in initData")
	}

	// Декодируем URL-encoded JSON
	decoded, err := url.QueryUnescape(userStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode user data: %w", err)
	}

	// Парсим JSON пользователя
	var user TelegramUser
	if err := json.Unmarshal([]byte(decoded), &user); err != nil {
		return nil, fmt.Errorf("failed to parse user JSON: %w", err)
	}

	return &user, nil
}

// CheckAuthDate проверяет, не истекла ли авторизация (24 часа)
func CheckAuthDate(initData string) (bool, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return false, fmt.Errorf("failed to parse initData: %w", err)
	}

	authDateStr := values.Get("auth_date")
	if authDateStr == "" {
		return false, fmt.Errorf("auth_date not found in initData")
	}

	var authDate int64
	fmt.Sscanf(authDateStr, "%d", &authDate)

	// Проверяем, что auth_date не старше 24 часов
	authTime := time.Unix(authDate, 0)
	now := time.Now()
	diff := now.Sub(authTime)

	return diff < 24*time.Hour, nil
}
