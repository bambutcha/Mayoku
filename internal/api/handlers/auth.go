package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Chelaran/mayoku/internal/models"
	"github.com/Chelaran/mayoku/internal/utils"

	logger "github.com/Chelaran/yagalog"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db        *gorm.DB
	botToken  string
	jwtSecret string
	logger    *logger.Logger
}

func NewAuthHandler(db *gorm.DB, botToken, jwtSecret string) *AuthHandler {
	log, _ := logger.NewLogger()
	return &AuthHandler{
		db:        db,
		botToken:  botToken,
		jwtSecret: jwtSecret,
		logger:    log,
	}
}

type AuthRequest struct {
	InitData string `json:"init_data"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

// HandleAuth обрабатывает POST /api/auth
func (h *AuthHandler) HandleAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.InitData == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "init_data is required"})
		return
	}

	// Проверяем подпись
	if h.botToken == "" {
		h.logger.Error("Bot token not configured")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Bot token not configured"})
		return
	}

	h.logger.Debug("Verifying initData signature")
	valid, err := utils.VerifyTelegramInitData(req.InitData, h.botToken)
	if err != nil {
		h.logger.Error("Failed to verify initData: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to verify initData: " + err.Error()})
		return
	}
	if !valid {
		h.logger.Warning("Invalid signature for initData")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid signature"})
		return
	}
	h.logger.Info("InitData signature verified successfully")

	// Проверяем срок действия (24 часа)
	validDate, err := utils.CheckAuthDate(req.InitData)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to check auth date: " + err.Error()})
		return
	}
	if !validDate {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Auth data expired"})
		return
	}

	// Парсим данные пользователя
	tgUser, err := utils.ParseTelegramUser(req.InitData)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to parse user data: " + err.Error()})
		return
	}

	// Находим или создаем пользователя
	var user models.User
	result := h.db.Where("tg_id = ?", tgUser.ID).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		// Создаем нового пользователя
		user = models.User{
			TgID:     tgUser.ID,
			Username: tgUser.Username,
		}
		if err := h.db.Create(&user).Error; err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create user"})
			return
		}
	} else if result.Error != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	} else {
		// Обновляем данные существующего пользователя
		user.Username = tgUser.Username
		if tgUser.PhotoURL != "" {
			user.AvatarURL = tgUser.PhotoURL
		}
		h.db.Save(&user)
	}

	// Генерируем JWT токен
	token, err := utils.GenerateJWT(user.ID, user.TgID, h.jwtSecret)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to generate token"})
		return
	}

	// Возвращаем токен и пользователя
	response := AuthResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
