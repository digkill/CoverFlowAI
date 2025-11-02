# CoverFlow AI Backend

Backend сервис для генерации обложек YouTube видео с помощью AI (Nano Banana Edit или OpenAI DALL-E 3).

## Требования

- Go 1.21 или выше
- Redis (для временного хранения изображений)
- API ключ от Nano Banana или OpenAI

## Установка

1. Установите зависимости:
```bash
go mod download
```

2. Создайте файл `.env`:
```bash
touch .env
```

3. Убедитесь, что Redis запущен:
```bash
# macOS
brew install redis
brew services start redis

# Linux
sudo apt-get install redis-server
sudo systemctl start redis

# Docker
docker run -d -p 6379:6379 redis
```

4. Добавьте API ключи в `.env`:
```
# Nano Banana API (обязательно)
NANO_BANANA_API_KEY=your_nano_banana_api_key_here

# OpenAI API (опционально, как альтернатива)
OPENAI_API_KEY=your_openai_api_key_here

# Google OAuth (для авторизации)
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/callback

# Session secret (для безопасности сессий)
SESSION_SECRET=your_random_secret_key_here

# Redis configuration (опционально, по умолчанию localhost:6379)
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

# Порт сервера
PORT=8080

# Публичный URL сервера (обязательно для работы с Nano Banana)
BASE_URL=http://localhost:8080

# Frontend URL (для редиректа после авторизации)
FRONTEND_URL=http://localhost:3000

# Lava Top Payment Integration
LAVA_SHOP_ID=your_lava_shop_id
LAVA_SECRET_KEY=your_lava_secret_key
LAVA_API_URL=https://api.lava.top  # Optional, defaults to this
```

**Важно:**
- **Redis обязателен** - используется для временного хранения изображений перед отправкой в API
- Изображения сохраняются в Redis на 30 минут, затем автоматически удаляются после генерации
- Сгенерированные обложки сохраняются локально в `storage/userid/` директории
- Для Nano Banana требуется публичный `BASE_URL` (используйте ngrok для локальной разработки)

**Настройка Google OAuth:**
1. Перейдите в [Google Cloud Console](https://console.cloud.google.com/)
2. Создайте новый проект или выберите существующий
3. Включите Google+ API
4. Создайте OAuth 2.0 credentials:
   - Тип: Web application
   - Authorized redirect URIs: `http://localhost:8080/api/auth/callback` (для production укажите ваш домен)
5. Скопируйте Client ID и Client Secret в `.env` файл

## Запуск

```bash
go run main.go
```

Сервер запустится на порту 8080 (или на порту, указанном в переменной окружения PORT).

## Запуск в Docker

Для контейнерного запуска используется `Dockerfile`, основанный на `golang:latest`.

1. Скопируйте `.env.example` в `.env` и заполните значения.
2. Соберите образ:
   ```bash
   docker build -t coverflow-backend .
   ```
3. Запустите контейнер (пример с пробросом порта и .env-файлом):
   ```bash
   docker run --env-file .env \
     -p 8080:8080 \
     -v $(pwd)/data:/srv/app/data \
     -v $(pwd)/storage:/srv/app/storage \
     coverflow-backend
   ```

Контейнер создаст каталоги `data/` и `storage/` внутри, они проброшены во внешние тома для сохранения базы SQLite и обложек. Redis и другие зависимости должны быть доступны контейнеру по адресу, указанному в `.env`.

## API Endpoints

### GET /api/health
Проверка здоровья сервиса.

### POST /api/generate-cover
Генерация обложки на основе коллажа.

**Request Body:**
```json
{
  "image": "data:image/png;base64,...",
  "provider": "nanobanana" // или "openai" (по умолчанию "nanobanana")
}
```

**Response:**
```json
{
  "id": "uuid",
  "image_url": "https://..."
}
```

## Провайдеры

### Nano Banana Edit (по умолчанию)
- Использует модель `google/nano-banana-edit`
- Преобразует коллаж в профессиональную обложку YouTube
- Формат: PNG, 16:9 для YouTube
- Максимальный размер входного изображения: 10MB
- Процесс: коллаж → Redis → публичный URL → Nano Banana API → результат → `storage/userid/`
- Кеш Redis автоматически очищается после генерации

### OpenAI DALL-E 3
- Использует модель `dall-e-3`
- Генерирует новую обложку на основе описания коллажа
- Формат: PNG, 1024x1024

## Структура проекта

- `main.go` - основной файл сервера с API endpoints и интеграцией с AI провайдерами
- `storage/` - директория для сохранения сгенерированных обложек (структура: `storage/userid/filename.png`)
- Redis - используется для временного хранения изображений коллажей (TTL: 30 минут)

## API Endpoints

### Авторизация
- `GET /api/auth/google` - начать OAuth поток (редирект на Google)
- `GET /api/auth/callback` - обработка callback от Google OAuth
- `GET /api/auth/me` - получить информацию о текущем пользователе
- `POST /api/auth/logout` - выйти из системы

### Генерация
- `GET /api/image/:imageId` - получить изображение из Redis кеша (используется Nano Banana API)
- `POST /api/generate-cover` - сгенерировать обложку
- `GET /storage/*` - статический доступ к сохраненным обложкам
