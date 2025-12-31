package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Chelaran/mayoku/internal/api"
	"github.com/Chelaran/mayoku/internal/config"
	"github.com/Chelaran/mayoku/internal/database"
	"github.com/Chelaran/mayoku/internal/game"
	"github.com/Chelaran/mayoku/internal/models"

	logger "github.com/Chelaran/yagalog"
)

func main() {
	// Инициализация логгера
	log, err := logger.NewLogger()
	if err != nil {
		panic(err)
	}

	// Загрузка конфигурации
	cfg, err := config.LoadFromFile(".env")
	if err != nil {
		cfg, err = config.Load()
		if err != nil {
			log.Fatal("Failed to load config: %v", err)
		}
	}

	log.Info("Configuration loaded successfully")

	// Подключение к PostgreSQL
	db, err := database.ConnectPostgres(cfg)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL: %v", err)
	}
	log.Info("PostgreSQL connected successfully")

	// Автомиграция моделей
	err = db.AutoMigrate(
		&models.User{},
		&models.Deck{},
		&models.Location{},
		&models.GameHistory{},
	)
	if err != nil {
		log.Fatal("Failed to run migrations: %v", err)
	}
	log.Info("Database migrations completed")

	// Подключение к Redis
	redisClient, err := database.ConnectRedis(cfg)
	if err != nil {
		log.Fatal("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Info("Redis connected successfully")

	// Подключение к MinIO
	minioClient, err := database.ConnectMinIO(cfg)
	if err != nil {
		log.Fatal("Failed to connect to MinIO: %v", err)
	}
	log.Info("MinIO connected successfully, bucket '%s' ready", cfg.MinIO.BucketName)

	// Создание Game Hub
	gameHub := game.NewHub(db, redisClient)
	log.Info("Game Hub initialized")

	// Создание HTTP сервера с Chi роутером
	addr := fmt.Sprintf("%s:%s", cfg.App.Host, cfg.App.Port)
	router := api.Router(api.RouterConfig{
		DB:          db,
		BotToken:    cfg.Telegram.BotToken,
		JWTSecret:   cfg.JWT.Secret,
		MinIO:       minioClient,
		MinIOBucket: cfg.MinIO.BucketName,
		Redis:       redisClient,
		GameHub:     gameHub,
	})

	server := &http.Server{
		Addr:    addr,
		Handler: router,
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
