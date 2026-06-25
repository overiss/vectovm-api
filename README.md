# vectovm-api

BFF и доменный API для платформы **VectoVM**. Сервис стоит между React-фронтендом (через nginx) и внешними компонентами: [go-oauthv2](https://github.com/overiss/go-oauthv2), PostgreSQL, HashiCorp Vault и [vectovm-mapi](https://github.com/overiss/vectovm-mapi).

## Архитектура

```
React → nginx → vectovm-api ──► PostgreSQL (connection storage)
                    │
                    ├── go-oauthv2 (регистрация, OAuth, JWKS)
                    ├── Vault API (master key для envelope encryption)
                    └── vectovm-mapi (datanode provisioning)
```

| Компонент | Роль |
|-----------|------|
| **go-oauthv2** | Хранит пользователей и выдаёт OAuth access/refresh tokens |
| **vectovm-api** | Доменная логика, связь user ↔ datanode ↔ VM |
| **PostgreSQL** | `users`, `datanodes`, `vms` |
| **Vault (API domain)** | Master key для envelope encryption SSH-учётных данных VM |
| **vectovm-mapi** | Асинхронный provisioning datanode по SSH |

## HTTP-сервер

Один listener на порту из `HTTP_PORT` (по умолчанию `:8081`). Маршруты без авторизации — signup, OAuth, health, swagger. Остальные — с `Authorization: Bearer`.

## Быстрый старт

### Требования

- Go 1.25+
- PostgreSQL
- HashiCorp Vault (master key)
- Запущенные go-oauthv2 и vectovm-mapi

### 1. Конфигурация

```bash
cp .env.example .env
# отредактируйте переменные окружения
```

### 2. Master key в Vault

```bash
vault kv put secret/vectovm-api/credentials master_key="$(openssl rand -base64 32)"
```

### 3. Запуск

```bash
make build
./bin/vectovm-api
```

Swagger UI:

```
http://localhost:8081/swagger/index.html
```

## Конфигурация

Основные переменные (полный список — в `.env.example`):

| Переменная | Описание |
|------------|----------|
| `HTTP_PORT` | Listen address (`:8081`) |
| `HTTP_NAME` | Имя сервера в логах |
| `POSTGRES_DSN` | PostgreSQL connection string |
| `OAUTH_ISSUER` | URL go-oauthv2 |
| `OAUTH_CLIENT_ID` / `OAUTH_CLIENT_SECRET` | OAuth client для BFF |
| `OAUTH_REDIRECT_URI` | Callback URL |
| `MAPI_BASE_URL` | URL vectovm-mapi (HTTPS) |
| `MAPI_BEARER_TOKEN` | Bearer token для mapi |
| `VAULT_ADDRESS` / `VAULT_TOKEN` | Vault API domain |
| `VAULT_CREDENTIALS_KEY_PATH` | Путь к master key (`secret/data/vectovm-api/credentials`) |

## Флоу пользователя

1. **Регистрация** — `POST /api/v1/signup` → go-oauthv2 + запись `oauth_user_id` и per-user DEK в PostgreSQL.
2. **Логин** — фронтенд → OAuth authorize на go-oauthv2 → `POST /api/v1/auth/token` (code + PKCE).
3. **Datanode** — `POST /api/v1/datanodes` с SSH-данными → vectovm-mapi (секреты в mapi-vault).
4. **VM** — `POST /api/v1/vms` с указанием `datanode_name` → проверка владения datanode → envelope encryption SSH login/password в PostgreSQL.

## Envelope encryption (VM credentials)

Схема аналогична AWS KMS:

```
Vault master_key
      ↓ wrap
users.encrypted_dek          ← симметричный DEK пользователя
      ↓ encrypt
vms.encrypted_credentials    ← AES-GCM(JSON {user, password})
```

- Master key **никогда** не хранится в БД.
- У каждого пользователя свой DEK в колонке `users.encrypted_dek`.
- SSH login/password VM **не возвращаются** в API-ответах.

## API

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| `GET` | `/healthz` | — | Liveness |
| `GET` | `/readyz` | — | Readiness (PostgreSQL) |
| `GET` | `/swagger/index.html` | — | Swagger UI |
| `POST` | `/api/v1/signup` | — | Регистрация |
| `POST` | `/api/v1/auth/token` | — | Обмен code → tokens |
| `POST` | `/api/v1/auth/refresh` | — | Refresh token |
| `POST` | `/api/v1/auth/logout` | — | Revoke refresh token |
| `GET` | `/api/v1/me` | Bearer | Профиль пользователя |
| `POST` | `/api/v1/datanodes` | Bearer | Создать datanode (async) |
| `GET` | `/api/v1/datanodes` | Bearer | Список datanodes |
| `POST` | `/api/v1/datanodes/vault/deploy` | Bearer | Deploy Vault |
| `GET` | `/api/v1/datanodes/jobs/:id` | Bearer | Статус job |
| `GET` | `/api/v1/datanodes/:name/runtime` | Bearer | Runtime info |
| `POST` | `/api/v1/vms` | Bearer | Создать VM |
| `GET` | `/api/v1/vms` | Bearer | Список VM |
| `GET` | `/api/v1/vms/:name` | Bearer | Получить VM |

### Пример: создать VM

```bash
curl -X POST http://localhost:8081/api/v1/vms \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-app-vm",
    "datanode_name": "primary-datanode",
    "host": "10.0.0.10",
    "port": 22,
    "ssh_user": "ubuntu",
    "ssh_password": "secret"
  }'
```

## Swagger / OpenAPI

Интерактивная документация: `http://localhost:8081/swagger/index.html`

Перегенерация после изменения аннотаций в handlers:

```bash
make swagger
```

Артефакты:

- `api/docs/swagger.json`
- `api/docs/swagger.yaml`
- `api/docs/docs.go`

## Разработка

```bash
make tidy      # go mod tidy
make swagger   # перегенерировать OpenAPI
make build     # собрать bin/vectovm-api
make run       # собрать и запустить
```

### Структура проекта

```
cmd/vectovm-api/          # entrypoint + swagger meta
api/docs/                 # сгенерированный OpenAPI
internal/
  app/                    # wiring, routes
  auth/                   # JWKS verifier
  client/oauth/           # go-oauthv2 client
  client/mapi/            # vectovm-mapi client
  crypto/                 # envelope encryption
  vault/                  # Vault keystore
  storage/postgres/       # миграции, пул
  repository/             # users, datanodes, vms
  service/                # auth, user, datanode, vm
  server/http/            # Gin handlers, middleware
```

## Связанные репозитории

- [go-oauthv2](https://github.com/overiss/go-oauthv2) — OAuth 2.0 authorization server
- [vectovm-mapi](https://github.com/overiss/vectovm-mapi) — internal management API для datanode

## License

Proprietary — VectoVM / Overiss.
