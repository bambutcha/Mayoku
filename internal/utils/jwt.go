package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims представляет claims для JWT токена
type JWTClaims struct {
	UserID uint  `json:"user_id"`
	TgID   int64 `json:"tg_id"`
	jwt.RegisteredClaims
}

// GenerateJWT создает JWT токен для пользователя
func GenerateJWT(userID uint, tgID int64, secret string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		TgID:   tgID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)), // 30 дней
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWT проверяет и парсит JWT токен
func ValidateJWT(tokenString, secret string) (*JWTClaims, error) {
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

