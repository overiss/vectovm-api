# vectovm-api

BFF and domain API for the **VectoVM** platform. The service sits between the React frontend (via nginx) and external components: [go-oauthv2](https://github.com/overiss/go-oauthv2), PostgreSQL, HashiCorp Vault, and [vectovm-mapi](https://github.com/overiss/vectovm-mapi).

## Architecture

```
React → nginx → vectovm-api ──► PostgreSQL (connection storage)
                    │
                    ├── go-oauthv2 (registration, OAuth, JWKS)
                    ├── Vault API (master key for envelope encryption)
                    └── vectovm-mapi (datanode provisioning)
```

| Component | Role |
|-----------|------|
| **go-oauthv2** | Stores users and issues OAuth access/refresh tokens |
| **vectovm-api** | Domain logic, user ↔ datanode ↔ VM relationships |
| **PostgreSQL** | `users`, `datanodes`, `vms` |
| **Vault (API domain)** | Master key for envelope encryption of VM SSH credentials |
| **vectovm-mapi** | Async datanode provisioning over SSH |

## HTTP server

Single listener on the port from `HTTP_PORT` (default `:8081`). Unauthenticated routes: signup, OAuth, swagger. All others require `Authorization: Bearer`.

## Quick start

### Requirements

- Go 1.25+
- PostgreSQL
- HashiCorp Vault (master key)
- Running go-oauthv2 and vectovm-mapi

### 1. Configuration

```bash
cp .env.example .env
# edit environment variables
```

### 2. Master key in Vault

```bash
vault kv put secret/vectovm-api/credentials master_key="$(openssl rand -base64 32)"
```

### 3. Run

```bash
make build
./bin/vectovm-api
```

Swagger UI:

```
http://localhost:8081/swagger/index.html
```

## Configuration

Key variables (full list in `.env.example`):

| Variable | Description |
|----------|-------------|
| `HTTP_PORT` | Listen address (`:8081`) |
| `HTTP_NAME` | Server name in logs |
| `POSTGRES_DSN` | PostgreSQL connection string |
| `OAUTH_ISSUER` | go-oauthv2 URL |
| `OAUTH_CLIENT_ID` / `OAUTH_CLIENT_SECRET` | OAuth client for BFF |
| `OAUTH_REDIRECT_URI` | Callback URL |
| `MAPI_BASE_URL` | vectovm-mapi URL (HTTPS) |
| `MAPI_BEARER_TOKEN` | Bearer token for mapi |
| `VAULT_ADDRESS` / `VAULT_TOKEN` | Vault API domain |
| `VAULT_CREDENTIALS_KEY_PATH` | Path to master key (`secret/data/vectovm-api/credentials`) |

## User flow

1. **Registration** — `POST /api/v1/signup` → go-oauthv2 + store `oauth_user_id` and per-user DEK in PostgreSQL.
2. **Login** — frontend → OAuth authorize on go-oauthv2 → `POST /api/v1/auth/token` (code + PKCE).
3. **Datanode** — `POST /api/v1/datanodes` with SSH credentials → vectovm-mapi (secrets in mapi-vault).
4. **VM** — `POST /api/v1/vms` with `datanode_name` → verify datanode ownership → envelope-encrypt SSH login/password in PostgreSQL.

## Envelope encryption (VM credentials)

Scheme similar to AWS KMS:

```
Vault master_key
      ↓ wrap
users.encrypted_dek          ← per-user symmetric DEK
      ↓ encrypt
vms.encrypted_credentials    ← AES-GCM(JSON {user, password})
```

- Master key is **never** stored in the database.
- Each user has their own DEK in the `users.encrypted_dek` column.
- VM SSH login/password are **not returned** in API responses.

## API

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/swagger/index.html` | — | Swagger UI |
| `POST` | `/api/v1/signup` | — | Registration |
| `POST` | `/api/v1/auth/token` | — | Exchange code → tokens |
| `POST` | `/api/v1/auth/refresh` | — | Refresh token |
| `POST` | `/api/v1/auth/logout` | — | Revoke refresh token |
| `GET` | `/api/v1/me` | Bearer | User profile |
| `POST` | `/api/v1/datanodes` | Bearer | Create datanode (async) |
| `GET` | `/api/v1/datanodes` | Bearer | List datanodes |
| `POST` | `/api/v1/datanodes/vault/deploy` | Bearer | Deploy Vault |
| `GET` | `/api/v1/datanodes/jobs/:id` | Bearer | Job status |
| `GET` | `/api/v1/datanodes/:name/runtime` | Bearer | Runtime info |
| `POST` | `/api/v1/vms` | Bearer | Create VM |
| `GET` | `/api/v1/vms` | Bearer | List VMs |
| `GET` | `/api/v1/vms/:name` | Bearer | Get VM |

Base URL in examples: `http://localhost:8081`. For protected endpoints, use `ACCESS_TOKEN` from the `POST /api/v1/auth/token` response.

### Request examples

#### Swagger UI

```bash
open http://localhost:8081/swagger/index.html
# or
curl -s http://localhost:8081/swagger/index.html
```

#### Registration — `POST /api/v1/signup`

```bash
curl -X POST http://localhost:8081/api/v1/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "her-secure-password"
  }'
```

#### Exchange code → tokens — `POST /api/v1/auth/token`

```bash
curl -X POST http://localhost:8081/api/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{
    "code": "oauth-authorization-code",
    "code_verifier": "pkce-code-verifier"
  }'
```

#### Refresh token — `POST /api/v1/auth/refresh`

```bash
curl -X POST http://localhost:8081/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "oauth-refresh-token"
  }'
```

#### Logout — `POST /api/v1/auth/logout`

```bash
curl -X POST http://localhost:8081/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "oauth-refresh-token"
  }'
```

#### User profile — `GET /api/v1/me`

```bash
curl http://localhost:8081/api/v1/me \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

#### Create datanode — `POST /api/v1/datanodes`

```bash
curl -X POST http://localhost:8081/api/v1/datanodes \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "primary-datanode",
    "host": "10.0.0.5",
    "port": 22,
    "user": "root",
    "password": "secret"
  }'
```

#### List datanodes — `GET /api/v1/datanodes`

```bash
curl http://localhost:8081/api/v1/datanodes \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

#### Deploy Vault — `POST /api/v1/datanodes/vault/deploy`

```bash
curl -X POST http://localhost:8081/api/v1/datanodes/vault/deploy \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "datanode_name": "primary-datanode"
  }'
```

#### Job status — `GET /api/v1/datanodes/jobs/:id`

```bash
curl http://localhost:8081/api/v1/datanodes/jobs/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

#### Runtime info — `GET /api/v1/datanodes/:name/runtime`

```bash
curl http://localhost:8081/api/v1/datanodes/primary-datanode/runtime \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

#### Create VM — `POST /api/v1/vms`

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

#### List VMs — `GET /api/v1/vms`

```bash
curl http://localhost:8081/api/v1/vms \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

#### Get VM — `GET /api/v1/vms/:name`

```bash
curl http://localhost:8081/api/v1/vms/my-app-vm \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

## Swagger / OpenAPI

Interactive documentation: `http://localhost:8081/swagger/index.html`

Regenerate after changing handler annotations:

```bash
make swagger
```

Artifacts:

- `api/docs/swagger.json`
- `api/docs/swagger.yaml`
- `api/docs/docs.go`

## Development

```bash
make tidy      # go mod tidy
make swagger   # regenerate OpenAPI
make build     # build bin/vectovm-api
make run       # build and run
```

### Project structure

```
cmd/vectovm-api/          # entrypoint + swagger meta
api/docs/                 # generated OpenAPI
internal/
  app/                    # wiring, routes
  auth/                   # JWKS verifier
  client/oauth/           # go-oauthv2 client
  client/mapi/            # vectovm-mapi client
  crypto/                 # envelope encryption
  vault/                  # Vault keystore
  storage/postgres/       # migrations, pool
  repository/             # users, datanodes, vms
  service/                # auth, user, datanode, vm
  server/http/            # Gin handlers, middleware
```

## Related repositories

- [go-oauthv2](https://github.com/overiss/go-oauthv2) — OAuth 2.0 authorization server
- [vectovm-mapi](https://github.com/overiss/vectovm-mapi) — internal management API for datanodes

## License

Proprietary — VectoVM / Overiss.
