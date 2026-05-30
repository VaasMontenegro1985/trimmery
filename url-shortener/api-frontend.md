# API для frontend

Контракт backend для UI. При расхождении с кодом приоритет у [`internal/http/router/router.go`](internal/http/router/router.go).

Базовый URL API: `http://localhost:8080` (или значение `BASE_URL` без завершающего `/`).

Все JSON-ошибки:

```json
{ "error": { "code": "код_ошибки", "message": "Текст для UI" } }
```

Авторизация: заголовок `Authorization: Bearer <access_token>` (после register/login).

---

## Общий workflow

### Без авторизации

| Сценарий | Ручки |
|----------|--------|
| Проверка, что API жив | `GET /health` |
| Сократить URL анонимно | `POST /shorten` (без заголовка Authorization) |
| Переход по короткой ссылке | `GET /{code}` → редирект 302 |

Анонимная ссылка **не попадает** в личный кабинет: нет stats, edit, delete, QR через API.

### С авторизацией (JWT)

1. `POST /auth/register` или `POST /auth/login` → сохранить `access_token` и `expires_at`.
2. Дальше для защищённых ручек: `Authorization: Bearer <token>`.
3. `POST /shorten` **с** JWT — ссылка привязывается к пользователю и видна в `GET /api/links`.
4. Личный кабинет: список, статистика, QR, редактирование, удаление — только `/api/*` с JWT.

### Опциональная авторизация

`POST /shorten`: JWT не обязателен, но если заголовок `Authorization` **передан**, токен должен быть валидным. Иначе `401 unauthorized` (не «тихий» анонимный режим).

### Ключевые особенности для UI

- **Два типа коротких URL:** публичный редирект `GET /{code}` (корень сайта); API — префикс `/api/`.
- **`short_url`** в JSON — полный URL (`BASE_URL/code`). **`qr_url`** в списке — **относительный** путь (`/api/links/{code}/qr`); для запроса QR добавьте origin API.
- **Владелец:** stats и QR только для своих ссылок; чужие/анонимные → `403 link_forbidden`. `PATCH` / `DELETE` чужой или несуществующей ссылки → `404 link_not_found` (в БД обновление/удаление идёт с фильтром `user_id`).
- **Soft delete:** после `DELETE` редирект и API по коду дают `404`. Алиас можно занять снова.
- **Смена `code` (PATCH):** старый `short_url` перестаёт работать; обновите ссылки и QR в UI.
- **Код ссылки:** 1–32 символа, только `0-9`, `a-z`, `A-Z` (без `-`, `_`, кириллицы).
- **Даты:** ISO 8601 / RFC3339 в JSON (`created_at`, `expires_at`, `visited_at`).
- **QR:** ответ — бинарный PNG, не JSON. Удобно: `<img src=".../qr?size=256" />` с заголовком Authorization через `fetch` + blob URL, или отдельная кнопка «Скачать».
- **Refresh token / logout:** в API нет — при истечении JWT снова login.
- **CORS:** настраивается на стороне деплоя; для dev часто proxy Vite → backend.

---

## 1. Health check

Проверка доступности сервиса (мониторинг, splash screen).

| | |
|---|---|
| **Эндпоинт** | `/health` |
| **Метод** | `GET` |
| **Auth** | нет |

**Ответ `200`:**

```json
{ "status": "ok" }
```

---

## 2. Регистрация

Создание пользователя и выдача JWT.

| | |
|---|---|
| **Эндпоинт** | `/auth/register` |
| **Метод** | `POST` |
| **Auth** | нет |
| **Content-Type** | `application/json` |

**Тело:**

```json
{ "email": "user@example.com", "password": "strong-password" }
```

Пароль: 8–72 байта. Email — валидный формат.

**Ответ `201`:**

```json
{
  "access_token": "<jwt>",
  "token_type": "Bearer",
  "expires_at": "2026-05-30T12:00:00Z",
  "user": { "id": 1, "email": "user@example.com" }
}
```

**Ошибки:** `400` (`invalid_json`, `invalid_email`, `invalid_password`), `409` `email_already_exists`, `500` `internal_error`.

---

## 3. Вход

Аутентификация существующего пользователя.

| | |
|---|---|
| **Эндпоинт** | `/auth/login` |
| **Метод** | `POST` |
| **Auth** | нет |

**Тело / ответ:** как у регистрации, но ответ **`200`** (не `201`).

**Ошибки:** `400` (валидация), `401` `invalid_credentials`, `500`.

---

## 4. Сокращение URL

Создание короткой ссылки; с JWT — привязка к аккаунту.

| | |
|---|---|
| **Эндпоинт** | `/shorten` |
| **Метод** | `POST` |
| **Auth** | опционально (`Bearer`) |

**Тело:**

```json
{
  "url": "https://example.com/long/path",
  "custom_code": "MyAlias123"
}
```

`url` — обязателен, `http`/`https`, до 2048 символов. `custom_code` — опционально (алиас).

**Ответ `201`:**

```json
{
  "code": "a8KzP2",
  "short_url": "http://localhost:8080/a8KzP2",
  "original_url": "https://example.com/long/path"
}
```

**Ошибки:** `400` (`invalid_json`, `url_required`, `invalid_url`, `invalid_code`), `401` (битый JWT при переданном Authorization), `409` `code_already_exists`, `500`.

**UI:** после login всегда слать JWT на shorten, чтобы ссылки попадали в кабинет.

---

## 5. Редирект по короткому коду

Публичный переход на оригинальный URL; учитывается в статистике (`clicks_count`, `visits`).

| | |
|---|---|
| **Эндпоинт** | `/{code}` |
| **Метод** | `GET` |
| **Auth** | нет |

**Ответ `302`:** заголовок `Location: <original_url>`.

**Ошибки (JSON):** `400` `invalid_code`, `404` `link_not_found`, `500`.

**UI:** открывать в новой вкладке или копировать `short_url`; не вызывать через XHR без follow redirects, если нужен именно переход.

---

## 6. Список ссылок пользователя

Личный кабинет: все активные ссылки текущего пользователя.

| | |
|---|---|
| **Эндпоинт** | `/api/links` |
| **Метод** | `GET` |
| **Auth** | обязательно |

**Ответ `200`:**

```json
{
  "links": [
    {
      "code": "a8KzP2",
      "short_url": "http://localhost:8080/a8KzP2",
      "original_url": "https://example.com/...",
      "clicks_count": 2,
      "qr_url": "/api/links/a8KzP2/qr",
      "created_at": "2026-05-29T09:00:00Z"
    }
  ]
}
```

**Ошибки:** `401` `unauthorized`, `500`.

**UI:** пустой массив — «нет ссылок»; для QR использовать `origin + qr_url` или прямой fetch на `/api/links/{code}/qr`.

---

## 7. Статистика переходов

Детали кликов по одной своей ссылке.

| | |
|---|---|
| **Эндпоинт** | `/api/links/:code/stats` |
| **Метод** | `GET` |
| **Auth** | обязательно |

**Query:** `limit` — число последних визитов (1–100; по умолчанию **50**, если не передан или ≤0).

**Ответ `200`:**

```json
{
  "code": "a8KzP2",
  "short_url": "http://localhost:8080/a8KzP2",
  "original_url": "https://example.com/...",
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

`ip` / `user_agent` могут быть `null`.

**Ошибки:** `401`, `400` (`invalid_code`, `invalid_limit`), `404` `link_not_found`, `403` `link_forbidden`, `500`.

---

## 8. QR-код (PNG)

Скачать/показать QR с закодированным `short_url`.

| | |
|---|---|
| **Эндпоинт** | `/api/links/:code/qr` |
| **Метод** | `GET` |
| **Auth** | обязательно |

**Query:** `size` — сторона в px (128–1024; по умолчанию **256**).

**Ответ `200`:** тело — **бинарный PNG**, `Content-Type: image/png`, `Content-Disposition: inline; filename="{code}.png"`, `Cache-Control: private, max-age=3600`.

**Ошибки (JSON):** `401`, `400` (`invalid_code`, `invalid_size`), `404`, `403`, `500`.

**UI:** `fetch` + `blob:` для `<img>`; или ссылка «Скачать» с тем же URL и Authorization.

---

## 9. Редактирование ссылки

Изменить оригинальный URL и/или алиас (`code`).

| | |
|---|---|
| **Эндпоинт** | `/api/links/:code` |
| **Метод** | `PATCH` |
| **Auth** | обязательно |

`:code` — текущий код в URL.

**Тело** (хотя бы одно поле):

```json
{
  "original_url": "https://example.com/new",
  "code": "NewAlias42"
}
```

**Ответ `200`:** объект ссылки (без `qr_url` в ответе PATCH — для QR пересоберите путь или обновите список):

```json
{
  "code": "NewAlias42",
  "short_url": "http://localhost:8080/NewAlias42",
  "original_url": "https://example.com/new",
  "clicks_count": 2,
  "created_at": "2026-05-29T09:00:00Z"
}
```

**Ошибки:** `401`, `400` (`invalid_json`, `nothing_to_update`, `invalid_url`, `invalid_code`), `404`, `409` `code_already_exists`, `500`.

---

## 10. Удаление ссылки

Soft delete: ссылка исчезает из списка и редиректа.

| | |
|---|---|
| **Эндпоинт** | `/api/links/:code` |
| **Метод** | `DELETE` |
| **Auth** | обязательно |

**Ответ `204`:** пустое тело.

**Ошибки:** `401`, `400` `invalid_code`, `404` `link_not_found` (нет ссылки, уже удалена или не ваша), `500`.

---

## Сводка: auth по ручкам

| Ручка | Auth |
|-------|------|
| `GET /health` | — |
| `POST /auth/register`, `POST /auth/login` | — |
| `POST /shorten` | опционально |
| `GET /{code}` | — |
| `GET /api/links` | JWT |
| `GET /api/links/:code/stats` | JWT |
| `GET /api/links/:code/qr` | JWT |
| `PATCH /api/links/:code` | JWT |
| `DELETE /api/links/:code` | JWT |

---

## Рекомендуемые экраны UI

1. **Гость:** форма shorten → показать `short_url` + кнопка «Копировать»; подсказка «Войдите, чтобы хранить ссылки».
2. **Login / Register** → сохранить token (localStorage / memory).
3. **Кабинет:** `GET /api/links` → таблица: original, short, clicks, действия (stats, QR, edit, delete).
4. **Stats:** модалка с `clicks_count` + таблица `visits`.
5. **QR:** preview через blob или download.
6. **Edit:** PATCH; при смене code предупредить, что старый URL недействителен.

---

## Коды ошибок (справочник)

| code | HTTP | Когда |
|------|------|--------|
| `unauthorized` | 401 | Нет/невалидный JWT на защищённой ручке |
| `invalid_credentials` | 401 | Неверный login |
| `invalid_json` | 400 | Тело не JSON |
| `url_required` | 400 | Нет `url` в shorten |
| `invalid_url` | 400 | URL не http(s) или слишком длинный |
| `invalid_code` | 400 | Недопустимый `code` / `:code` в path |
| `invalid_limit` | 400 | stats: limit вне 1–100 |
| `invalid_size` | 400 | qr: size вне 128–1024 |
| `nothing_to_update` | 400 | PATCH без полей |
| `link_not_found` | 404 | Нет активной ссылки |
| `link_forbidden` | 403 | Не владелец (stats, qr) |
| `code_already_exists` | 409 | Алиас занят |
| `email_already_exists` | 409 | Register |
| `not_found` | 404 | Несуществующий маршрут |
| `internal_error` | 500 | Сбой сервера |
