# CoverFlow AI

Сервис для создания и генерации обложек YouTube видео с помощью AI.

## Структура проекта

- `frontend/` - React приложение с редактором коллажей
- `backend/` - Go сервер с интеграцией Nano Banana Edit и OpenAI

## Быстрый старт

### Frontend

1. Перейдите в директорию frontend:
```bash
cd frontend
```

2. Установите зависимости:
```bash
npm install
```

3. Запустите dev сервер:
```bash
npm run dev
```

Frontend будет доступен на http://localhost:3000

### Backend

1. Перейдите в директорию backend:
```bash
cd backend
```

2. Установите зависимости:
```bash
go mod download
```

3. Создайте файл `.env`:
```bash
touch .env
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
SESSION_SECRET=your_random_secret_key_here

# Redis configuration (опционально, по умолчанию localhost:6379)
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

# Порт сервера
PORT=8080

# Публичный URL (обязательно для работы с Nano Banana)
BASE_URL=http://localhost:8080
FRONTEND_URL=http://localhost:3000
```

**Получение API ключей:**
- **Nano Banana**: Посетите [https://kie.ai/api-key](https://kie.ai/api-key), создайте аккаунт и получите API ключ

**Настройка Google OAuth:**
1. Перейдите в [Google Cloud Console](https://console.cloud.google.com/)
2. Создайте проект и включите Google+ API
3. Создайте OAuth 2.0 credentials (Web application)
4. Добавьте redirect URI: `http://localhost:8080/api/auth/callback`
5. Скопируйте Client ID и Client Secret в `.env`

**Установка Redis:**
- **macOS**: `brew install redis && brew services start redis`
- **Linux**: `sudo apt-get install redis-server && sudo systemctl start redis`
- **Docker**: `docker run -d -p 6379:6379 redis`

5. Запустите сервер:
```bash
go run main.go
```

Backend будет доступен на http://localhost:8080

## Использование

1. Откройте frontend в браузере (http://localhost:3000)
2. Загрузите изображения для создания коллажа
3. Расположите изображения на холсте (перетаскивание, изменение размера)
4. Выберите провайдера (Nano Banana или OpenAI) в выпадающем списке
5. Нажмите "Сгенерировать обложку"
6. Дождитесь завершения генерации (может занять несколько минут для Nano Banana)
7. Получите AI-сгенерированную обложку и скачайте её

## Технологии

### Frontend
- React 18
- TypeScript
- Vite
- Tailwind CSS
- shadcn/ui
- Konva.js (для редактора коллажей)

### Backend
- Go 1.21+
- Gin (HTTP framework)
- Nano Banana Edit API (основной провайдер)
- OpenAI API (DALL-E 3, альтернативный провайдер)

