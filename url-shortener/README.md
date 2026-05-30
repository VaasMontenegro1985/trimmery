# URL Shortener

Учебный сервис для сокращения URL: публичный shorten и редирект, JWT-авторизация, личный кабинет, статистика переходов, генераций QR-кодов

## Стек

- Go 1.26.2
- Gin
- PostgreSQL
- pgx
- goose
- JWT (`github.com/golang-jwt/jwt/v5`)
- QR: `github.com/skip2/go-qrcode`

## Переменные окружения
- `.env.example` - переменные приложения и Docker Compose;
- `.env.goose.example` - переменные для `goose`.

Создать локальные env-файлы:

```bash
cp .env.example .env
cp .env.goose.example .env.goose
```

После копирования нужно заполнить значения в `.env` и `.env.goose` под свое окружение.

Пример структуры `.env.goose.example`:

```env
GOOSE_DRIVER=postgres
GOOSE_DBSTRING=CHANGE_ME_POSTGRES_DSN
GOOSE_MIGRATION_DIR=./migrations
GOOSE_TABLE=goose_migrations
```

Пример структуры `.env.example`:

```bash
HTTP_ADDR=CHANGE_ME_HTTP_ADDR
BASE_URL=CHANGE_ME_BASE_URL
LOG_LEVEL=CHANGE_ME_LOG_LEVEL
JWT_SECRET=CHANGE_ME_AT_LEAST_32_CHARACTERS_SECRET
JWT_ACCESS_TTL=24h

DB_HOST=CHANGE_ME_DB_HOST
DB_PORT=CHANGE_ME_DB_PORT
DB_USER=CHANGE_ME_DB_USER
DB_PASSWORD=CHANGE_ME_DB_PASSWORD
DB_NAME=CHANGE_ME_DB_NAME
DB_SSLMODE=CHANGE_ME_DB_SSLMODE

POSTGRES_PORT=CHANGE_ME_POSTGRES_PUBLIC_PORT
POSTGRES_DB=CHANGE_ME_POSTGRES_DB
POSTGRES_USER=CHANGE_ME_POSTGRES_USER
POSTGRES_PASSWORD=CHANGE_ME_POSTGRES_PASSWORD
```

`HTTP_ADDR` задает адрес HTTP-сервера. `BASE_URL` используется при формировании короткой ссылки. `JWT_SECRET` используется для подписи JWT и должен быть не короче 32 символов. `JWT_ACCESS_TTL` задает время жизни access token, например `24h`. `DB_*` используются для сборки PostgreSQL DSN. Если задан `DATABASE_URL`, приложение использует его вместо отдельных `DB_*` параметров. `POSTGRES_*` использует Docker Compose. `LOG_LEVEL` управляет уровнем логирования.

## Запуск с нуля

### 1. Проверить зависимости

Нужны:

- Go;
- Docker и Docker Compose;
- `goose`.

Установить `goose`:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Проверить, что `goose` доступен:

```bash
goose -version
```

### 2. Запустить PostgreSQL

```bash
docker compose up -d postgres
```

Compose берет параметры PostgreSQL из локального `.env`:

```text
POSTGRES_PORT
POSTGRES_DB
POSTGRES_USER
POSTGRES_PASSWORD
```

Проверить, что контейнер запущен:

```bash
docker compose ps
```

### 3. Загрузить переменные окружения для миграций

`goose` читает настройки из переменных окружения. Загрузить `.env.goose` в текущую shell-сессию:

```bash
set -a
source .env.goose
set +a
```

Проверить, что переменные подхватились:

```bash
echo "$GOOSE_DBSTRING"
```

### 4. Применить миграции

Миграции в [`migrations/`](migrations/) — по одной таблице (финальная схема без последующих `ALTER`):

| Файл | Таблица |
|------|---------|
| `001_create_users.sql` | `users` |
| `002_create_links.sql` | `links` (+ индексы, partial unique на `code`) |
| `003_create_visits.sql` | `visits` |

```bash
goose up
```

Проверить статус:

```bash
goose status
```

После этого в БД должны быть таблицы `users`, `links`, `visits`.

Если раньше применялись старые спринтовые миграции (4 файла), сбросьте схему и накатите заново:

```bash
goose reset   # или DROP SCHEMA public CASCADE; CREATE SCHEMA public;
goose up
```

### 5. Подтянуть Go-зависимости

```bash
go mod tidy
```

### 6. Запустить приложение

Перед запуском приложения загрузить локальный `.env`:

```bash
set -a
source .env
set +a
go run ./cmd
```

## Проверка health endpoint

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ:

```json
{
  "status": "ok"
}
```

## Создание короткой ссылки

```bash
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/some/long/path?from=practice"}'
```

Ожидаемый ответ:

```json
{
  "code": "<some_awesome_code>",
  "short_url": "http://localhost:8080/a8KzP2",
  "original_url": "https://example.com/some/long/path?from=practice"
}
```

Если передать валидный JWT в заголовке `Authorization`, ссылка будет привязана к текущему пользователю. Без JWT ссылка создается в анонимном режиме.

```bash
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{"url":"https://example.com/some/long/path?from=practice"}'
```

### Кастомное имя (алиас) при создании

Можно указать желаемый короткий код через `custom_code`:

```bash
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/some/long/path?from=practice","custom_code":"MyAlias123"}'
```

Если `custom_code` занят активной ссылкой, сервис вернёт `409` с ошибкой `code_already_exists`.

## Регистрация пользователя

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"strong-password"}'
```

Ожидаемый ответ:

```json
{
  "access_token": "<jwt>",
  "token_type": "Bearer",
  "expires_at": "2026-05-30T12:00:00Z",
  "user": {
    "id": 1,
    "email": "user@example.com"
  }
}
```

## Авторизация пользователя

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"strong-password"}'
```

Ожидаемый ответ того же формата, что и регистрация, со статусом **`200 OK`** (не `201`).

## Список ссылок пользователя

```bash
curl http://localhost:8080/api/links \
  -H "Authorization: Bearer <access_token>"
```

Ожидаемый ответ:

```json
{
  "links": [
    {
      "code": "a8KzP2",
      "short_url": "http://localhost:8080/a8KzP2",
      "original_url": "https://example.com/some/long/path?from=practice",
      "clicks_count": 2,
      "qr_url": "/api/links/a8KzP2/qr",
      "created_at": "2026-05-29T09:00:00Z"
    }
  ]
}
```

Поле `qr_url` — относительный путь к эндпоинту QR; изображение не встраивается в JSON списка.

## Скачать QR-код ссылки (только владелец)

Эндпоинт: `GET /api/links/:code/qr` (JWT required).

Возвращает PNG (`image/png`). QR кодирует короткий URL (`short_url`). Уровень коррекции: Medium. Размер стороны задаётся query `size` (по умолчанию `256`, диапазон `128`–`1024`).

```bash
curl -o qr.png "http://localhost:8080/api/links/<code>/qr?size=256" \
  -H "Authorization: Bearer <access_token>"
```

После soft delete ссылка недоступна (`404 link_not_found`). Чужие и анонимные ссылки — `403 link_forbidden`.

## Редактирование и удаление ссылок (только владелец)

### Обновить ссылку

Эндпоинт: `PATCH /api/links/:code` (JWT required).

Можно менять `original_url`, `code` (алиас) или оба поля.

```bash
curl -X PATCH http://localhost:8080/api/links/<code> \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{"original_url":"https://example.com/new","code":"NewAlias42"}'
```

Важно: при смене `code` **старый короткий URL перестаёт работать** (будет `404 link_not_found`), потому что редирект зависит от текущего значения `code`.

Если новый `code` занят активной ссылкой — `409 code_already_exists`.

### Удалить ссылку (soft delete)

Эндпоинт: `DELETE /api/links/:code` (JWT required).

```bash
curl -X DELETE http://localhost:8080/api/links/<code> \
  -H "Authorization: Bearer <access_token>"
```

После удаления:

- редирект по коду возвращает `404 link_not_found`;
- статистика и QR по ссылке через API недоступны (`404 link_not_found`);
- алиас освобождается и может быть использован снова (partial unique index по `code` только для `deleted_at IS NULL`).

## Статистика переходов по ссылке

Сделать несколько переходов по короткой ссылке (публично, без JWT):

```bash
curl -i http://localhost:8080/<code>
curl -i http://localhost:8080/<code>
```

Получить статистику по ссылке (только владелец):

```bash
curl "http://localhost:8080/api/links/<code>/stats?limit=50" \
  -H "Authorization: Bearer <access_token>"
```

Ожидаемый ответ:

```json
{
  "code": "a8KzP2",
  "short_url": "http://localhost:8080/a8KzP2",
  "original_url": "https://example.com/some/long/path?from=practice",
  "clicks_count": 2,
  "visits": [
    {
      "visited_at": "2026-05-29T10:15:00Z",
      "ip": "203.0.113.10",
      "user_agent": "Mozilla/5.0 ..."
    }
  ]
}
```

## Проверка редиректа

Подставить `code` из ответа `POST /shorten`:

```bash
curl -i http://localhost:8080/<code>
```

Ожидаемый результат:

```http
HTTP/1.1 302 Found
Location: https://example.com/some/long/path?from=practice
```

Проверка с автоматическим переходом:

```bash
curl -L http://localhost:8080/<code>
```

Если код не найден:

```json
{
  "error": {
    "code": "link_not_found",
    "message": "Short link not found"
  }
}
```