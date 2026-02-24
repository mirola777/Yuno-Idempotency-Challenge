# Yuno Idempotency Challenge

## Project

Go backend service implementing idempotency for payment operations using hexagonal architecture.

## Stack

- Go 1.24
- Echo v4 (HTTP framework)
- GORM + PostgreSQL (persistence)
- Docker Compose (infrastructure)

## Structure

```
internal/
  domain/         Pure models, errors, port interfaces
  application/    Use cases, service logic, DI container
  infrastructure/ Database repos, payment processor simulator
  presentation/   Echo HTTP server, handlers, middleware
utils/
  config/         .env loader with OS fallback
  fingerprint/    SHA256 request hashing
```

## Running

```
docker-compose up -d
go run main.go
```

## Testing

```
go test ./...
```

## Conventions

- No comments in code
- No emojis
- All dependencies injected through interfaces defined in domain/ports.go
- container.go wires all adapters
- Domain errors carry HTTP status codes
- Echo custom error handler maps AppError to JSON
