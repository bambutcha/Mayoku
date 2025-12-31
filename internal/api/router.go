package api

import (
	"net/http"

	"github.com/Chelaran/mayoku/internal/api/handlers"
	"github.com/Chelaran/mayoku/internal/api/middleware"
	"github.com/Chelaran/mayoku/internal/game"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RouterConfig содержит зависимости для роутера
type RouterConfig struct {
	DB          *gorm.DB
	BotToken    string
	JWTSecret   string
	MinIO       *minio.Client
	MinIOBucket string
	Redis       *redis.Client
	GameHub     *game.Hub
}

// Router настраивает маршруты приложения
func Router(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.URLFormat)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Главная страница для Mini App
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		html := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Mayoku - Spy Game</title>
    <script src="https://telegram.org/js/telegram-web-app.js"></script>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 16px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            max-width: 600px;
            width: 100%;
        }
        h1 {
            color: #333;
            margin-bottom: 20px;
            font-size: 24px;
        }
        button {
            background: #3390ec;
            color: white;
            border: none;
            padding: 14px 28px;
            border-radius: 10px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            margin: 10px 5px;
            width: 100%;
            transition: background 0.3s;
        }
        button:hover {
            background: #2878c4;
        }
        .success {
            background: #4caf50;
        }
        .success:hover {
            background: #45a049;
        }
        pre {
            background: #f5f5f5;
            padding: 15px;
            border-radius: 10px;
            overflow-x: auto;
            font-size: 13px;
            border: 2px solid #e0e0e0;
            margin: 15px 0;
            max-height: 300px;
            overflow-y: auto;
            word-break: break-all;
        }
        .info {
            background: #e3f2fd;
            padding: 15px;
            border-radius: 10px;
            margin: 15px 0;
            border-left: 4px solid #3390ec;
            font-size: 14px;
        }
        .success-box {
            background: #e8f5e9;
            border-left-color: #4caf50;
        }
        .error-box {
            background: #ffebee;
            border-left-color: #f44336;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🕵️ Mayoku - Spy Game</h1>
        
        <div class="info">
            <strong>Получить InitData:</strong><br>
            Нажмите кнопку ниже, чтобы получить initData для тестирования API
        </div>

        <button onclick="showInitData()">📋 Получить initData</button>
        <button onclick="testAPI()" class="success">🚀 Тест API</button>
        <button onclick="copyToClipboard()">📎 Копировать</button>

        <div id="status"></div>
        <pre id="output">Нажмите кнопку для получения initData...</pre>
    </div>

    <script>
        const tg = window.Telegram.WebApp;
        tg.ready();
        tg.expand();

        function showInitData() {
            const initData = tg.initData;
            const initDataUnsafe = tg.initDataUnsafe;
            const status = document.getElementById('status');
            const output = document.getElementById('output');
            
            if (!initData) {
                status.innerHTML = '<div class="info error-box"><strong>❌ Ошибка:</strong> initData недоступен.</div>';
                output.textContent = 'initData не найден';
                return;
            }

            status.innerHTML = '<div class="info success-box"><strong>✅ Успешно!</strong> initData получен.</div>';
            
            const data = {
                initData: initData,
                initDataUnsafe: initDataUnsafe,
                version: tg.version,
                platform: tg.platform
            };
            
            output.textContent = JSON.stringify(data, null, 2);
            window.lastInitData = initData;
        }

        async function testAPI() {
            const initData = tg.initData || window.lastInitData;
            if (!initData) {
                alert('Сначала получите initData!');
                return;
            }

            const status = document.getElementById('status');
            const output = document.getElementById('output');
            status.innerHTML = '<div class="info">⏳ Отправка запроса к API...</div>';

            try {
                const response = await fetch('/api/auth', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ init_data: initData })
                });

                let data;
                const text = await response.text();
                
                try {
                    data = JSON.parse(text);
                } catch (e) {
                    // Если не JSON, показываем как текст
                    data = { error: text };
                }
                
                if (response.ok) {
                    status.innerHTML = '<div class="info success-box"><strong>✅ Успешно!</strong> Авторизация прошла. JWT токен получен.</div>';
                    output.textContent = JSON.stringify(data, null, 2);
                    window.authToken = data.token;
                } else {
                    status.innerHTML = '<div class="info error-box"><strong>❌ Ошибка:</strong> ' + response.status + ' ' + response.statusText + '</div>';
                    output.textContent = JSON.stringify(data, null, 2);
                }
            } catch (error) {
                status.innerHTML = '<div class="info error-box"><strong>❌ Ошибка:</strong> ' + error.message + '</div>';
                console.error('API Error:', error);
            }
        }

        function copyToClipboard() {
            const output = document.getElementById('output');
            const text = output.textContent;
            
            if (navigator.clipboard) {
                navigator.clipboard.writeText(text).then(() => {
                    const status = document.getElementById('status');
                    status.innerHTML = '<div class="info success-box">✅ Скопировано в буфер обмена!</div>';
                });
            } else {
                const textarea = document.createElement('textarea');
                textarea.value = text;
                document.body.appendChild(textarea);
                textarea.select();
                document.execCommand('copy');
                document.body.removeChild(textarea);
                
                const status = document.getElementById('status');
                status.innerHTML = '<div class="info success-box">✅ Скопировано в буфер обмена!</div>';
            }
        }

        window.addEventListener('load', () => {
            setTimeout(showInitData, 500);
        });
    </script>
</body>
</html>`
		w.Write([]byte(html))
	})

	// Инициализация handlers
	authHandler := handlers.NewAuthHandler(cfg.DB, cfg.BotToken, cfg.JWTSecret)
	userHandler := handlers.NewUserHandler(cfg.DB)
	uploadHandler := handlers.NewUploadHandler(cfg.MinIO, cfg.MinIOBucket)
	deckHandler := handlers.NewDeckHandler(cfg.DB)
	gameHandler := handlers.NewGameHandler(cfg.GameHub, cfg.DB)
	wsHandler := handlers.NewWebSocketHandler(cfg.GameHub, cfg.DB)
	adminHandler := handlers.NewAdminHandler(cfg.DB)

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Auth routes (публичные)
		r.Post("/auth", authHandler.HandleAuth)

		// Protected routes (требуют JWT)
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(cfg.JWTSecret))

			// User routes
			r.Get("/user/me", userHandler.HandleGetMe)

			// Decks routes
			r.Route("/decks", func(r chi.Router) {
				r.Post("/", deckHandler.HandleCreateDeck)       // POST /api/decks - создание набора
				r.Get("/", deckHandler.HandleGetDecks)          // GET /api/decks - список наборов
				r.Get("/{id}", deckHandler.HandleGetDeck)       // GET /api/decks/:id - получение набора
				r.Put("/{id}", deckHandler.HandleUpdateDeck)    // PUT /api/decks/:id - обновление набора
				r.Delete("/{id}", deckHandler.HandleDeleteDeck) // DELETE /api/decks/:id - удаление набора
			})

			// Upload routes
			r.Route("/upload", func(r chi.Router) {
				r.Post("/", uploadHandler.HandleUpload)                  // POST /api/upload - загрузка картинки в MinIO
				r.Get("/presigned", uploadHandler.HandleGetPresignedURL) // GET /api/upload/presigned - получение presigned URL
			})

			// Game routes
			r.Route("/game", func(r chi.Router) {
				r.Post("/rooms", gameHandler.HandleCreateRoom) // POST /api/game/rooms - создание комнаты
				r.Get("/rooms", gameHandler.HandleListRooms)   // GET /api/game/rooms - список комнат
				r.Get("/ws", wsHandler.HandleWebSocket)        // GET /api/game/ws - WebSocket подключение
			})

			// Admin routes (требуют админских прав)
			r.Route("/admin", func(r chi.Router) {
				r.Use(middleware.AdminMiddleware(cfg.DB))

				// Deck moderation
				r.Get("/decks/pending", adminHandler.HandleGetPendingDecks)  // GET /api/admin/decks/pending - колоды на модерации
				r.Get("/decks", adminHandler.HandleGetAllDecks)              // GET /api/admin/decks - все колоды
				r.Put("/decks/{id}/approve", adminHandler.HandleApproveDeck) // PUT /api/admin/decks/:id/approve - одобрить
				r.Put("/decks/{id}/reject", adminHandler.HandleRejectDeck)   // PUT /api/admin/decks/:id/reject - отклонить
			})
		})
	})

	return r
}
