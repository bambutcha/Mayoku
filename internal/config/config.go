package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

// Config представляет конфигурацию приложения
type Config struct {
	App struct {
		Host string `env:"APP_HOST" env-default:"0.0.0.0"`
		Port string `env:"APP_PORT" env-default:"8080"`
	}

	Postgres struct {
		Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
		Port     string `env:"POSTGRES_PORT" env-default:"5432"`
		User     string `env:"POSTGRES_USER" env-default:"postgres"`
		Password string `env:"POSTGRES_PASSWORD" env-default:"postgres"`
		DBName   string `env:"POSTGRES_DB" env-default:"mayoku"`
		SSLMode  string `env:"POSTGRES_SSLMODE" env-default:"disable"`
	}

	Redis struct {
		Host     string `env:"REDIS_HOST" env-default:"localhost"`
		Port     string `env:"REDIS_PORT" env-default:"6379"`
		Password string `env:"REDIS_PASSWORD" env-default:""`
		DB       int    `env:"REDIS_DB" env-default:"0"`
	}

	MinIO struct {
		Endpoint        string `env:"MINIO_ENDPOINT" env-default:"localhost:9000"`
		AccessKeyID     string `env:"MINIO_ACCESS_KEY_ID" env-default:"minioadmin"`
		SecretAccessKey string `env:"MINIO_SECRET_ACCESS_KEY" env-default:"minioadmin"`
		UseSSL          bool   `env:"MINIO_USE_SSL" env-default:"false"`
		BucketName      string `env:"MINIO_BUCKET_NAME" env-default:"mayoku"`
	}

	Telegram struct {
		BotToken string `env:"TELEGRAM_BOT_TOKEN" env-default:""`
	}
}

// Load загружает конфигурацию из .env файла
func Load() (*Config, error) {
	var cfg Config

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadFromFile загружает конфигурацию из указанного файла
func LoadFromFile(path string) (*Config, error) {
	var cfg Config

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

