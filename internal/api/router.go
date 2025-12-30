package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Router настраивает маршруты приложения
func Router() http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Auth routes
		r.Route("/auth", func(r chi.Router) {
			// POST /api/auth - валидация Telegram initData
			// TODO: реализовать
		})

		// User routes
		r.Route("/user", func(r chi.Router) {
			// GET /api/user/me - профиль текущего пользователя
			// TODO: реализовать
		})

		// Decks routes
		r.Route("/decks", func(r chi.Router) {
			// POST /api/decks - создание набора
			// GET /api/decks - список наборов
			// TODO: реализовать
		})

		// Upload routes
		r.Route("/upload", func(r chi.Router) {
			// POST /api/upload - загрузка картинки в MinIO
			// TODO: реализовать
		})
	})

	return r
}
