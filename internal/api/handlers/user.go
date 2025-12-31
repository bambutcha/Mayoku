package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Chelaran/mayoku/internal/api/middleware"
	"github.com/Chelaran/mayoku/internal/models"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{
		db: db,
	}
}

// HandleGetMe обрабатывает GET /api/user/me
func (h *UserHandler) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем user_id из контекста (добавлен middleware)
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем пользователя из БД
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Подсчитываем количество созданных наборов (если не синхронизировано)
	var decksCount int64
	h.db.Model(&models.Deck{}).Where("author_id = ?", userID).Count(&decksCount)
	if user.DecksCreated != int(decksCount) {
		user.DecksCreated = int(decksCount)
		h.db.Save(&user)
	}

	// Формируем ответ со статистикой
	response := map[string]interface{}{
		"id":             user.ID,
		"tg_id":          user.TgID,
		"username":       user.Username,
		"avatar_url":     user.AvatarURL,
		"created_at":     user.CreatedAt,
		"updated_at":     user.UpdatedAt,
		"is_admin":       user.IsAdmin,
		"is_super_admin": user.IsSuperAdmin,
		"statistics": map[string]interface{}{
			"games_played":  user.GamesPlayed,
			"wins_spy":      user.WinsSpy,
			"wins_local":    user.WinsLocal,
			"losses_spy":    user.LossesSpy,
			"losses_local":  user.LossesLocal,
			"decks_created": user.DecksCreated,
			"win_rate":      calculateWinRate(user),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// calculateWinRate вычисляет общий винрейт
func calculateWinRate(user models.User) float64 {
	totalWins := user.WinsSpy + user.WinsLocal
	totalGames := user.GamesPlayed
	if totalGames == 0 {
		return 0.0
	}
	return float64(totalWins) / float64(totalGames) * 100.0
}
