package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Chelaran/mayoku/internal/utils"
)

// AuthMiddleware проверяет JWT токен из заголовка Authorization
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Извлекаем токен из заголовка Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Проверяем формат "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token := parts[1]

			// Валидируем токен
			claims, err := utils.ValidateJWT(token, jwtSecret)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Сохраняем claims в контекст
			r = r.WithContext(contextWithUserID(r.Context(), claims.UserID, claims.TgID))

			next.ServeHTTP(w, r)
		})
	}
}

// contextKey - тип для ключей контекста
type contextKey string

const (
	userIDKey contextKey = "user_id"
	tgIDKey   contextKey = "tg_id"
)

// contextWithUserID добавляет user_id и tg_id в контекст
func contextWithUserID(ctx context.Context, userID uint, tgID int64) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)
	ctx = context.WithValue(ctx, tgIDKey, tgID)
	return ctx
}

// GetUserID извлекает user_id из контекста
func GetUserID(ctx context.Context) (uint, bool) {
	userID, ok := ctx.Value(userIDKey).(uint)
	return userID, ok
}

// GetTgID извлекает tg_id из контекста
func GetTgID(ctx context.Context) (int64, bool) {
	tgID, ok := ctx.Value(tgIDKey).(int64)
	return tgID, ok
}
