package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Chelaran/mayoku/internal/config"

	logger "github.com/Chelaran/yagalog"
)

func main() {
	// Инициализация логгера
	log, err := logger.NewLogger()
	if err != nil {
		panic(err)
	}

	// Загрузка конфигурации
	// Сначала пробуем загрузить из .env файла, если не получается - из переменных окружения
	cfg, err := config.LoadFromFile(".env")
	if err != nil {
		// Пробуем загрузить из переменных окружения
		cfg, err = config.Load()
		if err != nil {
			log.Fatal("Failed to load config: %v", err)
		}
	}

	log.Info("Configuration loaded successfully")

	// Создание HTTP сервера
	addr := fmt.Sprintf("%s:%s", cfg.App.Host, cfg.App.Port)
	mux := http.NewServeMux()

	// Простой health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		log.Info("Starting HTTP server on %s", addr)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start: %v", err)
		}
	}()

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited")
}
