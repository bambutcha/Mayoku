# Локальная разработка с Telegram Mini App

## Проблема

BotFather требует **HTTPS URL** для Mini App, но `localhost` не поддерживает HTTPS и не доступен извне.

## Решение 1: Использование CloudPub (рекомендуется для России)

См. подробную инструкцию: [CLOUDPUB_SETUP.md](./CLOUDPUB_SETUP.md)

**Быстро:**
1. Зарегистрируйтесь на [cloudpub.ru](https://cloudpub.ru)
2. Создайте публикацию: `http://localhost:8080` → получите HTTPS URL
3. Используйте этот URL в BotFather

## Решение 2: Использование ngrok

### Шаг 1: Установите ngrok

```bash
# Linux
wget https://bin.equinox.io/c/bNyj1mQVY4c/ngrok-v3-stable-linux-amd64.tgz
tar -xzf ngrok-v3-stable-linux-amd64.tgz
sudo mv ngrok /usr/local/bin/

# Или через snap
sudo snap install ngrok

# Или через пакетный менеджер
# Ubuntu/Debian
sudo apt install ngrok
```

### Шаг 2: Запустите ngrok туннель

```bash
# Запустите ваш сервер на localhost:8080
docker compose up -d

# Создайте туннель
ngrok http 8080
```

Вы получите URL вида: `https://abc123.ngrok-free.app`

### Шаг 3: Используйте ngrok URL в BotFather

1. Скопируйте HTTPS URL из ngrok (например: `https://abc123.ngrok-free.app`)
2. Отправьте его в BotFather при создании Mini App
3. BotFather примет этот URL

### Шаг 4: Откройте Mini App

1. Найдите вашего бота `@MayokuBot` в Telegram
2. Откройте меню бота
3. Нажмите на Mini App
4. Mini App откроется через ngrok туннель

## Альтернативные решения

### Вариант 1: Использовать временный публичный URL

Можно использовать другие сервисы:
- **localtunnel**: `npx localtunnel --port 8080`
- **serveo**: `ssh -R 80:localhost:8080 serveo.net`
- **Cloudflare Tunnel**: более сложная настройка, но бесплатная

### Вариант 2: Развернуть на временном хостинге

Для тестирования можно использовать:
- **Vercel** (бесплатно)
- **Netlify** (бесплатно)
- **Railway** (бесплатный tier)
- **Render** (бесплатно)

### Вариант 3: Пропустить Mini App и использовать тестовый скрипт

Если нужно только получить initData для тестирования API:

1. Создайте простой HTML файл с Telegram Web App SDK
2. Загрузите его на любой хостинг (даже GitHub Pages)
3. Откройте через Telegram Mini App
4. Получите initData из консоли

## Быстрая настройка с ngrok

```bash
# 1. Установите ngrok (если еще не установлен)
# См. инструкции выше

# 2. Запустите приложение
docker compose up -d

# 3. Создайте туннель
ngrok http 8080

# 4. Скопируйте HTTPS URL (например: https://abc123.ngrok-free.app)

# 5. В BotFather отправьте этот URL при создании Mini App

# 6. Откройте Mini App в Telegram и получите initData
```

## Важные замечания

1. **ngrok бесплатный план** имеет ограничения:
   - URL меняется при каждом перезапуске
   - Ограничение по трафику
   - Может быть медленнее

2. **Для продакшена** используйте:
   - Постоянный домен
   - SSL сертификат (Let's Encrypt бесплатно)
   - Надежный хостинг

3. **Безопасность**:
   - Не используйте ngrok URL в продакшене
   - Регулярно меняйте JWT_SECRET
   - Используйте HTTPS везде

## Проверка работы

После настройки ngrok:

```bash
# Проверьте, что туннель работает
curl https://your-ngrok-url.ngrok-free.app/health

# Должен вернуть: OK
```

## Обновление URL в BotFather

Если ngrok URL изменился:

1. В BotFather отправьте `/myapps`
2. Выберите ваше приложение
3. Отправьте `/editapp`
4. Выберите "Web App URL"
5. Введите новый ngrok URL

