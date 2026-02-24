# Yuno Idempotency Challenge

## Project

Go backend service implementing idempotency for payment operations using hexagonal architecture.

## Stack

- Go 1.24
- Echo v4 (HTTP framework)
- GORM + PostgreSQL (persistence)
- Docker Compose (infrastructure)
- Air (hot reload for development)

## Structure

```
internal/
  domain/              Pure models, errors (with localization), port interfaces
  application/         Use cases, DI container (handles DB connection and migrations)
  infrastructure/
    gorm/              GORM repositories, connection, migrations (package gormdb)
    processor/         Payment processor simulator
  presentation/        Echo HTTP server, handlers, middleware
  utils/
    config/            Environment-aware config (APP_ENV: dev/test/prod)
    fingerprint/       SHA256 request hashing
```

## Running

Development with Air (hot reload):

```
docker compose up -d
air
```

Development fully containerized:

```
docker compose --profile dev up --build
```

Production:

```
docker compose --profile prod up --build
```

## Testing

```
APP_ENV=test go test ./...
```

## Environment

`APP_ENV` controls the environment mode: `dev` (default), `test`, `prod`. The config provides `IsDev()`, `IsProd()`, `IsTest()` helpers.

## Conventions

- No comments in code
- No emojis
- All dependencies injected through interfaces defined in domain/ports.go
- container.go wires all adapters, handles DB connection and migrations internally
- main.go is fully agnostic: only imports config, container, and server
- Domain errors carry HTTP status codes with per-language translations (Messages map)
- Each error factory in payment.go embeds its own Messages with en/es translations
- AppError.Localize(lang) returns a localized copy of the error
- Echo custom error handler maps AppError to JSON with Accept-Language support
