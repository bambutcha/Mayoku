# 🔐 Настройка администратора

## Способ 1: Через SQL (для первого админа)

Если у вас еще нет ни одного администратора, используйте SQL запрос:

```sql
-- Назначить админом пользователя по ID
UPDATE users SET is_admin = true WHERE id = 1;

-- Или по Telegram ID
UPDATE users SET is_admin = true WHERE tg_id = 666535426;
```

### Через Docker Compose

```bash
# Подключиться к PostgreSQL контейнеру
docker-compose exec postgres psql -U postgres -d mayoku

# Выполнить запрос
UPDATE users SET is_admin = true WHERE id = 1;
```

## Способ 2: Через API (требует существующего админа)

Если у вас уже есть хотя бы один администратор:

### 1. Авторизуйтесь как админ
```bash
# Получите JWT токен через POST /api/auth
curl -X POST http://localhost:8080/api/auth \
  -H "Content-Type: application/json" \
  -d '{"init_data": "YOUR_TELEGRAM_INIT_DATA"}'
```

### 2. Назначьте нового админа
```bash
# Сделать пользователя админом (замените {user_id} и {token})
curl -X PUT http://localhost:8080/api/admin/users/{user_id}/make-admin \
  -H "Authorization: Bearer {your_jwt_token}"
```

### 3. Проверьте список админов
```bash
curl -X GET http://localhost:8080/api/admin/users/admins \
  -H "Authorization: Bearer {your_jwt_token}"
```

## Способ 3: Через миграцию (рекомендуется для продакшена)

Создайте SQL миграцию для первого админа:

```sql
-- migrations/001_add_first_admin.sql
-- Замените tg_id на ваш Telegram ID
INSERT INTO users (tg_id, username, is_admin, created_at, updated_at)
VALUES (666535426, 'admin', true, NOW(), NOW())
ON CONFLICT (tg_id) DO UPDATE SET is_admin = true;
```

## Проверка прав администратора

После назначения админа, проверьте доступ:

```bash
# Попробуйте получить список колод на модерации
curl -X GET http://localhost:8080/api/admin/decks/pending \
  -H "Authorization: Bearer {your_jwt_token}"
```

Если получили список колод - вы админ! ✅  
Если получили `403 Forbidden` - проверьте, что `is_admin = true` в БД.

## Управление админами

### Список всех админов
```bash
GET /api/admin/users/admins
```

### Назначить админа
```bash
PUT /api/admin/users/:id/make-admin
```

### Убрать права админа
```bash
PUT /api/admin/users/:id/remove-admin
```

## Важно

- Для первого админа используйте SQL запрос
- Все админские endpoints требуют JWT токен + `is_admin = true`
- Админ может назначать других админов через API

