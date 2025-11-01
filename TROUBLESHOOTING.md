# Решение проблем

## Ошибка 500: Failed to generate cover

### Причина
Чаще всего эта ошибка возникает из-за отсутствия `IMGBB_API_KEY` в файле `.env`.

### Решение

1. **Получите бесплатный API ключ ImageBB:**
   - Перейдите на https://api.imgbb.com/
   - Зарегистрируйтесь (бесплатно)
   - Создайте API ключ

2. **Добавьте ключ в `.env` файл backend:**
   ```bash
   cd backend
   echo "IMGBB_API_KEY=your_key_here" >> .env
   ```

3. **Перезапустите backend сервер:**
   ```bash
   go run main.go
   ```

### Другие возможные причины

#### Ошибка: "Nano Banana API key not configured"
- Проверьте, что `NANO_BANANA_API_KEY` установлен в `.env` файле
- Получите ключ на https://kie.ai/api-key

#### Ошибка: "authentication failed"
- Проверьте правильность API ключей
- Убедитесь, что ключи не содержат лишних пробелов

#### Ошибка: "insufficient account balance"
- Пополните баланс на аккаунте Nano Banana
- Проверьте баланс на https://kie.ai/

#### Ошибка: "rate limit exceeded"
- Превышен лимит запросов
- Подождите несколько минут и попробуйте снова

## Проверка настроек

Убедитесь, что в `backend/.env` файле установлены все необходимые ключи:

```env
NANO_BANANA_API_KEY=your_nano_banana_key
IMGBB_API_KEY=your_imgbb_key
PORT=8080
```

## Логи для отладки

Если проблема сохраняется, проверьте логи backend сервера. Они покажут:
- Успешность загрузки изображения на ImageBB
- Ошибки при создании задачи Nano Banana
- Статус опроса задачи

## Альтернативное решение

Если вы не можете использовать ImageBB, установите публичный URL в `BASE_URL`:

```env
BASE_URL=https://your-public-server.com
```

Используйте ngrok для создания публичного туннеля к localhost:
```bash
ngrok http 8080
# Используйте полученный URL в BASE_URL
```

