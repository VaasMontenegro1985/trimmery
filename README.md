# TRIMMERY

TRIMMERY — сервис сокращения ссылок и генерации QR-кодов.

Учебный full-stack проект по кейсу Т1: "Сервис для сокращения URL-адресов и генерации QR-кодов".

## Краткое описание

Проект позволяет:

- сокращать длинные URL;
- создавать кастомные alias;
- выполнять редирект по короткой ссылке;
- регистрироваться и входить в личный кабинет;
- просматривать свои ссылки;
- видеть базовую статистику переходов;
- получать QR-коды;
- редактировать и удалять ссылки.

## Стек

Frontend:

- React;
- Vite;
- CSS;
- библиотека `qrcode` как fallback для анонимных ссылок.

Backend:

- Go;
- Gin;
- PostgreSQL;
- Goose migrations;
- JWT auth;
- Docker Compose.

## Структура проекта

```text
t1/
  trimmery/       # frontend на React + Vite
  url-shortener/  # backend на Go + Gin + PostgreSQL
```

Frontend общается с backend по REST API. PostgreSQL поднимается через Docker Compose. Миграции backend выполняются через goose.

## Что реализовано

- `POST /shorten` — создание короткой ссылки;
- `GET /:code` — редирект по короткому коду;
- регистрация и вход пользователя;
- JWT-авторизация;
- dashboard / личный кабинет пользователя;
- `GET /api/links` — список ссылок пользователя;
- backend QR для авторизованных пользовательских ссылок;
- frontend fallback QR для анонимных ссылок;
- `clicks_count` — базовый счётчик переходов;
- `PATCH /api/links/:code` — редактирование ссылки;
- `DELETE /api/links/:code` — удаление ссылки;
- адаптивный интерфейс.

## Как запустить локально

### Backend

```bash
cd url-shortener
docker compose up -d postgres
set -a && source .env.goose && set +a
go run github.com/pressly/goose/v3/cmd/goose@latest up
set -a && source .env && set +a && go run ./cmd
```

Команды `set -a` и `source` рассчитаны на Git Bash, Linux или macOS. На Windows удобнее запускать их через Git Bash.

### Frontend

```bash
cd trimmery
npm install
npm run dev -- --host 127.0.0.1 --port 3000
```

Локальные URL:

- frontend: <http://127.0.0.1:3000>
- backend: <http://localhost:8080>

## Env-файлы

Реальные `.env` файлы не должны коммититься в git.

Примеры env-файлов:

- `url-shortener/.env.example`
- `url-shortener/.env.goose.example`
- `trimmery/.env.example`

Важные переменные backend:

- `HTTP_ADDR` — адрес запуска HTTP-сервера;
- `BASE_URL` — базовый URL backend для генерации коротких ссылок;
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` — подключение backend к PostgreSQL;
- `POSTGRES_PORT`, `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD` — параметры PostgreSQL для Docker Compose;
- `JWT_SECRET` — секрет для подписи JWT;
- `JWT_ACCESS_TTL` — время жизни access token.

Важные переменные goose:

- `GOOSE_DRIVER` — драйвер миграций;
- `GOOSE_DBSTRING` — DSN подключения к PostgreSQL;
- `GOOSE_MIGRATION_DIR` — папка с миграциями;
- `GOOSE_TABLE` — таблица goose.

Важные переменные frontend:

```env
VITE_API_BASE_URL=http://localhost:8080
```

## Основные API endpoints

Public:

- `GET /health` — проверка состояния backend;
- `POST /shorten` — создание короткой ссылки, анонимно или с JWT;
- `GET /:code` — редирект на оригинальный URL.

Auth:

- `POST /auth/register` — регистрация пользователя;
- `POST /auth/login` — вход пользователя.

User:

- `GET /api/links` — получить ссылки авторизованного пользователя;
- `GET /api/links/:code/qr?size=256` — получить PNG QR-код для пользовательской ссылки;
- `PATCH /api/links/:code` — изменить оригинальный URL или alias ссылки;
- `DELETE /api/links/:code` — удалить ссылку пользователя.

## Проверка backend

```bash
cd url-shortener
go test ./...
go build ./...
bash run-api-e2e.sh
```

## Проверка frontend

```bash
cd trimmery
npm run build
```

## Сценарий демонстрации

1. Открыть главную страницу.
2. Сократить ссылку без входа.
3. Перейти по short URL.
4. Зарегистрироваться или войти.
5. Создать ссылку с alias.
6. Открыть кабинет.
7. Проверить статистику.
8. Открыть QR-код.
9. Отредактировать ссылку.
10. Удалить ссылку.

## Статус проекта

Проект находится на стадии MVP. Локально реализован основной функционал. Публичный деплой выполняется отдельно.

## Важные замечания

- Для пользовательских ссылок QR загружается через backend endpoint с JWT.
- Для анонимных ссылок используется frontend fallback QR, так как защищённая QR-ручка требует авторизацию.
- `JWT_ACCESS_TTL` для локальной разработки можно ставить `24h`.
- `.env` и `.env.goose` не должны попадать в git.
