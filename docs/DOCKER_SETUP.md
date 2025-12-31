# 🐳 Docker Setup Guide

## Быстрый старт

### 1. Настройка переменных окружения

#### Для Docker Compose (.env в корне проекта)
**Обязательно!** Docker Compose использует `.env` из корня проекта для подстановки переменных.

Скопируйте `.env.example` в `.env` в корне проекта:
```bash
cp .env.example .env
```

Отредактируйте `.env` и укажите:
- `TELEGRAM_BOT_TOKEN` - токен вашего Telegram бота
- `JWT_SECRET` - секретный ключ для JWT (используйте сильный случайный ключ)

#### Для локальной разработки Frontend (.env в frontend/)
Если запускаете frontend отдельно (не через Docker), скопируйте:
```bash
cp frontend/.env.example frontend/.env
```

По умолчанию API URL: `http://localhost:8080`

#### Для локальной разработки Backend (backend/.env)
Если запускаете backend отдельно (не через Docker), скопируйте:
```bash
cp backend/.env.example backend/.env
```

### 2. Запуск через Docker Compose

```bash
docker compose up --build
```

Или в фоновом режиме:
```bash
docker compose up -d --build
```

### 3. Проверка работы

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379
- **MinIO Console**: http://localhost:9001 (minioadmin/minioadmin)
- **MinIO API**: http://localhost:9000

### 4. Остановка

```bash
docker compose down
```

Для удаления всех данных (volumes):
```bash
docker compose down -v
```

## Структура сервисов

### Frontend
- **Порт**: 3000
- **Технологии**: Next.js 15 (static export) + Nginx
- **Переменные окружения**: `NEXT_PUBLIC_API_URL` (устанавливается через build arg)
- **Зависимости**: Backend

### Backend
- **Порт**: 8080
- **Переменные окружения**: загружаются из `.env` в корне проекта
- **Зависимости**: PostgreSQL, Redis, MinIO

### PostgreSQL
- **Порт**: 5432
- **База данных**: mayoku
- **Пользователь**: postgres
- **Пароль**: postgres

### Redis
- **Порт**: 6379
- **Используется для**: хранения состояния игровых комнат

### MinIO
- **API порт**: 9000
- **Console порт**: 9001
- **Используется для**: хранения изображений колод и локаций

## Переменные окружения

### Backend (.env)
```env
# Application
APP_HOST=0.0.0.0
APP_PORT=8080

# PostgreSQL
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=mayoku
POSTGRES_SSLMODE=disable

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# MinIO
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY_ID=minioadmin
MINIO_SECRET_ACCESS_KEY=minioadmin
MINIO_USE_SSL=false
MINIO_BUCKET_NAME=mayoku

# Telegram
TELEGRAM_BOT_TOKEN=your_bot_token_here

# JWT
JWT_SECRET=change-me-in-production-use-strong-secret
```

### Frontend (frontend/.env)
```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Troubleshooting

### Backend не может подключиться к БД
- Убедитесь, что PostgreSQL запущен: `docker compose ps`
- Проверьте логи: `docker compose logs postgres`

### MinIO не создает bucket
- Зайдите в MinIO Console (http://localhost:9001)
- Создайте bucket `mayoku` вручную
- Или проверьте логи: `docker compose logs minio`

### Проблемы с переменными окружения
- Убедитесь, что `.env` файл находится в корне проекта
- Проверьте, что все переменные заполнены
- Перезапустите контейнеры: `docker compose restart backend`

