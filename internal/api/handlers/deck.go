package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Chelaran/mayoku/internal/api/middleware"
	"github.com/Chelaran/mayoku/internal/models"
	logger "github.com/Chelaran/yagalog"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type DeckHandler struct {
	db  *gorm.DB
	log *logger.Logger
}

func NewDeckHandler(db *gorm.DB) *DeckHandler {
	log, _ := logger.NewLogger()
	return &DeckHandler{
		db:  db,
		log: log,
	}
}

// CreateDeckRequest запрос на создание колоды
type CreateDeckRequest struct {
	Name      string `json:"name"`
	IsPublic  bool   `json:"is_public"`
	Locations []struct {
		Name     string   `json:"name"`
		ImageURL string   `json:"image_url"`
		Roles    []string `json:"roles"`
	} `json:"locations"`
}

// UpdateDeckRequest запрос на обновление колоды
type UpdateDeckRequest struct {
	Name      string            `json:"name,omitempty"`
	IsPublic  *bool             `json:"is_public,omitempty"`
	Status    models.DeckStatus `json:"status,omitempty"`
	Locations []struct {
		ID       uint     `json:"id,omitempty"`
		Name     string   `json:"name,omitempty"`
		ImageURL string   `json:"image_url,omitempty"`
		Roles    []string `json:"roles,omitempty"`
	} `json:"locations,omitempty"`
}

// HandleCreateDeck создает новую колоду
// POST /api/decks
func (h *DeckHandler) HandleCreateDeck(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	var req CreateDeckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}

	// Валидация
	if req.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Name is required"})
		return
	}

	// Создаем колоду
	deck := models.Deck{
		AuthorID: userID,
		Name:     req.Name,
		IsPublic: req.IsPublic,
		Status:   models.DeckStatusDraft, // По умолчанию черновик
	}

	// Создаем локации
	for _, locReq := range req.Locations {
		location := models.Location{
			DeckID:   0, // Будет установлено после создания deck
			Name:     locReq.Name,
			ImageURL: locReq.ImageURL,
			Roles:    models.StringArray(locReq.Roles),
		}
		deck.Locations = append(deck.Locations, location)
	}

	// Сохраняем в БД
	if err := h.db.Create(&deck).Error; err != nil {
		h.log.Error("Failed to create deck: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create deck"})
		return
	}

	h.log.Info("Deck created: %s (ID: %d, Author: %d)", deck.Name, deck.ID, userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(deck)
}

// HandleGetDecks возвращает список колод
// GET /api/decks?my=true&public=true&status=approved
func (h *DeckHandler) HandleGetDecks(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context()) // Может быть не авторизован для публичных колод

	var decks []models.Deck
	query := h.db

	// Фильтры
	my := r.URL.Query().Get("my")
	public := r.URL.Query().Get("public")
	status := r.URL.Query().Get("status")

	if my == "true" && userID > 0 {
		// Только мои колоды
		query = query.Where("author_id = ?", userID)
	} else if public == "true" {
		// Только публичные
		query = query.Where("is_public = ? AND status = ?", true, models.DeckStatusApproved)
	} else if userID > 0 {
		// Авторизованный пользователь видит свои + публичные
		query = query.Where("author_id = ? OR (is_public = ? AND status = ?)", userID, true, models.DeckStatusApproved)
	} else {
		// Неавторизованный видит только публичные
		query = query.Where("is_public = ? AND status = ?", true, models.DeckStatusApproved)
	}

	// Фильтр по статусу
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Загружаем с локациями
	if err := query.Preload("Locations").Preload("Author").Find(&decks).Error; err != nil {
		h.log.Error("Failed to get decks: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get decks"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decks)
}

// HandleGetDeck возвращает конкретную колоду
// GET /api/decks/:id
func (h *DeckHandler) HandleGetDeck(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	deckIDStr := chi.URLParam(r, "id")
	deckID, err := strconv.ParseUint(deckIDStr, 10, 32)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid deck ID"})
		return
	}

	var deck models.Deck
	query := h.db.Preload("Locations").Preload("Author")

	// Проверяем доступ
	if userID > 0 {
		// Авторизованный пользователь видит свои или публичные
		query = query.Where("id = ? AND (author_id = ? OR (is_public = ? AND status = ?))", deckID, userID, true, models.DeckStatusApproved)
	} else {
		// Неавторизованный видит только публичные
		query = query.Where("id = ? AND is_public = ? AND status = ?", deckID, true, models.DeckStatusApproved)
	}

	if err := query.First(&deck).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Deck not found"})
			return
		}
		h.log.Error("Failed to get deck: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get deck"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deck)
}

// HandleUpdateDeck обновляет колоду
// PUT /api/decks/:id
func (h *DeckHandler) HandleUpdateDeck(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	deckIDStr := chi.URLParam(r, "id")
	deckID, err := strconv.ParseUint(deckIDStr, 10, 32)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid deck ID"})
		return
	}

	// Проверяем, что колода принадлежит пользователю
	var deck models.Deck
	if err := h.db.Where("id = ? AND author_id = ?", deckID, userID).First(&deck).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Deck not found or access denied"})
			return
		}
		h.log.Error("Failed to get deck: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get deck"})
		return
	}

	var req UpdateDeckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}

	// Обновляем поля
	if req.Name != "" {
		deck.Name = req.Name
	}
	if req.IsPublic != nil {
		deck.IsPublic = *req.IsPublic
	}
	if req.Status != "" {
		// Пользователь может изменить статус только на Pending (для публикации)
		if req.Status == models.DeckStatusPending {
			deck.Status = models.DeckStatusPending
		} else if req.Status == models.DeckStatusDraft {
			deck.Status = models.DeckStatusDraft
		}
		// Approved и Rejected может устанавливать только админ (TODO)
	}

	// Обновляем локации, если они указаны
	if len(req.Locations) > 0 {
		// Удаляем старые локации
		h.db.Where("deck_id = ?", deck.ID).Delete(&models.Location{})

		// Создаем новые
		for _, locReq := range req.Locations {
			location := models.Location{
				DeckID:   deck.ID,
				Name:     locReq.Name,
				ImageURL: locReq.ImageURL,
				Roles:    models.StringArray(locReq.Roles),
			}
			if locReq.ID > 0 {
				location.ID = locReq.ID
			}
			deck.Locations = append(deck.Locations, location)
		}
	}

	// Сохраняем изменения
	if err := h.db.Save(&deck).Error; err != nil {
		h.log.Error("Failed to update deck: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update deck"})
		return
	}

	// Загружаем обновленные данные
	h.db.Preload("Locations").Preload("Author").First(&deck, deck.ID)

	h.log.Info("Deck updated: %s (ID: %d)", deck.Name, deck.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deck)
}

// HandleDeleteDeck удаляет колоду
// DELETE /api/decks/:id
func (h *DeckHandler) HandleDeleteDeck(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	deckIDStr := chi.URLParam(r, "id")
	deckID, err := strconv.ParseUint(deckIDStr, 10, 32)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid deck ID"})
		return
	}

	// Проверяем, что колода принадлежит пользователю
	var deck models.Deck
	if err := h.db.Where("id = ? AND author_id = ?", deckID, userID).First(&deck).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Deck not found or access denied"})
			return
		}
		h.log.Error("Failed to get deck: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get deck"})
		return
	}

	// Удаляем колоду (локации удалятся каскадно из-за constraint:OnDelete:CASCADE)
	if err := h.db.Delete(&deck).Error; err != nil {
		h.log.Error("Failed to delete deck: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete deck"})
		return
	}

	h.log.Info("Deck deleted: %s (ID: %d)", deck.Name, deck.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
