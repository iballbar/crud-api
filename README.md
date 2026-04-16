# testarise

CRUD API assignment using **Go + Gin + Gorm**, with **unit tests**, **Postgres**, and **Docker Compose**.

## Hexagonal structure

The project follows Hexagonal Architecture:

- `internal/domain` - entities and domain errors
- `internal/ports` - contracts for use cases, repositories, and caches
- `internal/application` - business logic and decorators (e.g., caching)
- `internal/adapters/http` - Gin handlers, router, and middlewares
- `internal/adapters/postgres` - Gorm-based persistence
- `internal/adapters/redis` - Redis-based caching
- `internal/config` - Type-safe environment configuration (caarlos0/env)
- `cmd/api` - dependency injection and application entry point
- `tests/integration` - E2E tests using Testcontainers (Postgres)
- `internal/db/migrate.go` - Database schema migrations

## Database design

This project implements CRUD for `users`.

### Table: `users`

- `id` (UUID, PK)
- `name` (TEXT, NOT NULL)
- `email` (TEXT, NOT NULL, UNIQUE)
- `created_at` (TIMESTAMPTZ)
- `updated_at` (TIMESTAMPTZ)

Gorm auto-migrates this schema at startup (see `internal/db/migrate.go`).

## Endpoints

- `GET /healthz`
- `POST /v1/users`
- `GET /v1/users?page=1&pageSize=20`
- `GET /v1/users/:id`
- `PUT /v1/users/:id`
- `DELETE /v1/users/:id`

## Performance & Observability
- **Structured Logging**: Uses `log/slog` for JSON (prod) or Text (dev) logging
- **Middleware**: Request ID tracking, recovery, and detailed access logs
- **Centralized Error Handling**: Unified mapping from domain errors to HTTP status codes
- **DB connection pooling**: Configurable `MaxOpenConns`, `MaxIdleConns`, etc.
- **Gorm prepared statements**: Enabled by default for performance
- **Redis cache**: Transparent caching via Decorator pattern (configurable TTL via `REDIS_USER_TTL`)
- **Health Checks**: Detailed `/healthz` verifying DB and Redis connectivity
- **pprof**: Available in non-production at `/debug/pprof/`

## Configuration

The project uses environment variables for configuration (see `.env.example`):

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_ENV` | environment (development/production) | `development` |
| `PORT` | HTTP server port | `8080` |
| `DATABASE_DSN` | Postgres connection string | (see .env.example) |
| `DB_MAX_OPEN_CONNS` | Max open DB connections | `25` |
| `REDIS_ADDR` | Redis host:port | (optional) |
| `REDIS_USER_TTL` | Cache duration for user data | `30s` |

## Development (Makefile)

The project includes a `Makefile` for common tasks:

- `make help` - show available commands
- `make tidy` - clean up dependencies
- `make test` - run all tests
- `make test-unit` - run unit tests only
- `make tools` - install dev tools (mockery)
- `make mocks` - regenerate mocks
- `make docker-up` - start the environment with Docker Compose
- `make docker-down` - stop the Docker environment
- `make docker-logs` - follow Docker logs

## Run locally (no Docker)

Set a Postgres DSN, then:

```bash
# Using Go directly
go test ./...
go run ./cmd/api

# Using Makefile
make tidy
go run ./cmd/api
```

## Run with Docker Compose

```bash
# Using Docker directly
docker compose up --build

# Using Makefile
make docker-up
```

API runs on `http://localhost:8080`.

## Unit tests

Tests use **Testify Suites** and **Mockery** for fast, hermetic testing of the application logic and HTTP handlers.

```bash
# Using Go directly
go test ./internal/application/... ./internal/adapters/http/... -count=1 -v

# Using Makefile
make test-unit
```

## Integration tests

Integration tests use **Testcontainers** to spin up a real Postgres instance for E2E validation of the persistence layer and API.

```bash
# Using Go directly
go test ./tests/integration/... -count=1 -v

# Using Makefile
make test-integration
```

