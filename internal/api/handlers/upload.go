package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Chelaran/mayoku/internal/api/middleware"
	logger "github.com/Chelaran/yagalog"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type UploadHandler struct {
	minioClient *minio.Client
	bucketName  string
	log         *logger.Logger
}

func NewUploadHandler(minioClient *minio.Client, bucketName string) *UploadHandler {
	log, _ := logger.NewLogger()
	return &UploadHandler{
		minioClient: minioClient,
		bucketName:  bucketName,
		log:         log,
	}
}

type UploadResponse struct {
	URL      string `json:"url"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
}

func (h *UploadHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	// Получаем user_id из контекста
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	// Проверяем метод
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// Парсим multipart form (максимум 10MB)
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		h.log.Error("Failed to parse multipart form: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to parse form"})
		return
	}

	// Получаем файл
	file, header, err := r.FormFile("file")
	if err != nil {
		h.log.Error("Failed to get file from form: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "File is required"})
		return
	}
	defer file.Close()

	// Проверяем тип файла (только изображения)
	contentType := header.Header.Get("Content-Type")
	allowedTypes := []string{"image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp"}
	isAllowed := false
	for _, t := range allowedTypes {
		if contentType == t {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Only image files are allowed (JPEG, PNG, GIF, WebP)"})
		return
	}

	// Генерируем уникальное имя файла
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		// Определяем расширение по Content-Type
		switch contentType {
		case "image/jpeg", "image/jpg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		default:
			ext = ".jpg"
		}
	}

	fileID := uuid.New().String()
	fileName := fmt.Sprintf("%s%s", fileID, ext)
	objectName := fmt.Sprintf("users/%d/%s", userID, fileName)

	// Загружаем в MinIO
	ctx := context.Background()
	_, err = h.minioClient.PutObject(ctx, h.bucketName, objectName, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		h.log.Error("Failed to upload file to MinIO: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to upload file"})
		return
	}

	// Формируем URL (для MinIO нужен публичный URL или через presigned URL)
	// В продакшене лучше использовать presigned URL или настроить публичный доступ
	fileURL := fmt.Sprintf("/%s/%s", h.bucketName, objectName)

	// Если MinIO настроен с публичным доступом, можно использовать полный URL
	// fileURL := fmt.Sprintf("http://%s/%s/%s", h.minioClient.EndpointURL().Host, h.bucketName, objectName)

	h.log.Info("File uploaded successfully: %s (user: %d, size: %d bytes)", objectName, userID, header.Size)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UploadResponse{
		URL:      fileURL,
		FileName: fileName,
		FileSize: header.Size,
	})
}

// HandleGetPresignedURL генерирует presigned URL для загрузки файла
func (h *UploadHandler) HandleGetPresignedURL(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	// Получаем параметры из query
	fileName := r.URL.Query().Get("file_name")
	if fileName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "file_name is required"})
		return
	}

	// Генерируем уникальное имя
	ext := filepath.Ext(fileName)
	if ext == "" {
		ext = ".jpg"
	}
	fileID := uuid.New().String()
	objectName := fmt.Sprintf("users/%d/%s%s", userID, fileID, ext)

	// Генерируем presigned URL (действителен 1 час)
	ctx := context.Background()
	presignedURL, err := h.minioClient.PresignedPutObject(ctx, h.bucketName, objectName, time.Hour)
	if err != nil {
		h.log.Error("Failed to generate presigned URL: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to generate upload URL"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"upload_url":  presignedURL.String(),
		"object_name": objectName,
	})
}
