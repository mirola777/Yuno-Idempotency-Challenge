# Yuno Idempotency Challenge

Backend service that implements idempotency for payment operations, preventing duplicate charges caused by network retries.

Built for the Meridian Rides scenario: a Southeast Asian ride-hailing platform operating across Indonesia, Thailand, Vietnam, and the Philippines where unreliable mobile networks cause payment request retries.

## Architecture

Hexagonal (Ports and Adapters) with four layers:

```
presentation --> application --> domain <-- infrastructure
```

| Layer | Responsibility |
|---|---|
| domain | Models, errors, port interfaces (zero external dependencies) |
| application | Business logic, idempotency engine, DI container |
| infrastructure | GORM repositories, payment processor simulator |
| presentation | Echo HTTP server, handlers, middleware |

See [docs/architecture.md](docs/architecture.md) for details.

## Quick Start

### Prerequisites

- Go 1.24+
- Docker and Docker Compose
- PostgreSQL 16 (provided via Docker Compose)

### Setup

```bash
cp .env.example .env
docker-compose up -d
go run main.go
```

The service starts on port 8080 by default.

### Without Docker

Point to an existing PostgreSQL instance by setting environment variables:

```bash
export DB_HOST=your-host
export DB_PORT=5432
export DB_USER=your-user
export DB_PASSWORD=your-password
export DB_NAME=your-db
go run main.go
```

## API Endpoints

| Method | Path | Description |
|---|---|---|
| POST | /v1/payments | Create payment (requires X-Idempotency-Key header) |
| GET | /v1/payments/:id | Get payment by ID |
| GET | /v1/idempotency/:key | Lookup by idempotency key |
| GET | /health | Health check |

See [docs/api.md](docs/api.md) for full reference with examples.

## Idempotency Flow

1. Client sends POST /v1/payments with X-Idempotency-Key header
2. Service computes SHA256 fingerprint of the request payload
3. Within a PostgreSQL transaction with SELECT FOR UPDATE:
   - First request: processes payment, stores result, returns 201
   - Duplicate request (same key + same payload): returns cached response
   - Conflicting request (same key + different payload): returns 409
4. Concurrent requests with the same key are serialized via row-level locking

See [docs/concurrency.md](docs/concurrency.md) for the concurrency strategy.

## Test Card Numbers

| Card Number | Outcome |
|---|---|
| 4111111111111111 | SUCCEEDED |
| 5500000000000004 | SUCCEEDED |
| 4000000000000002 | FAILED (insufficient_funds) |
| 4000000000000069 | FAILED (expired_card) |
| 4000000000000119 | FAILED (processing_error) |
| 4000000000000259 | PENDING |

## Testing

### Unit Tests

```bash
go test ./...
```

### Postman Collection

Import the collection and environment from `tests/postman/`:

1. Open Postman
2. Import `tests/postman/yuno_idempotency_challenge.postman_collection.json`
3. Import `tests/postman/yuno_idempotency_challenge.postman_environment.json`
4. Select the environment
5. Run the collection (Runner > Run Collection)

The collection contains 26 requests across 8 folders covering all acceptance criteria.

### Demo Script

```bash
chmod +x tests/scripts/demo.sh
./tests/scripts/demo.sh
```

Demonstrates: payment creation, idempotent retries, conflict detection, failed/pending payments, lookups, validation errors, and concurrent request handling.

## Configuration

| Variable | Default | Description |
|---|---|---|
| APP_PORT | 8080 | HTTP server port |
| DB_HOST | localhost | PostgreSQL host |
| DB_PORT | 5432 | PostgreSQL port |
| DB_USER | idempotency | Database user |
| DB_PASSWORD | idempotency123 | Database password |
| DB_NAME | idempotency_db | Database name |
| DB_SSLMODE | disable | PostgreSQL SSL mode |
| IDEMPOTENCY_KEY_TTL | 24h | Time before idempotency keys expire |
| CLEANUP_INTERVAL | 1h | Interval for expired key cleanup |
| GRACEFUL_TIMEOUT | 5s | Graceful shutdown timeout |

Configuration loads from `.env` file first, falls back to OS environment variables if `.env` is not present.

See [docs/infrastructure.md](docs/infrastructure.md) for infrastructure details.

## Project Structure

```
internal/
  application/payment/
    container.go          DI wiring
    service.go            Idempotency engine
    service_test.go       Unit tests
  domain/
    models.go             Payment, IdempotencyRecord, enums
    errors.go             Domain errors with HTTP codes
    ports.go              Repository, Processor, Service interfaces
  infrastructure/
    database/
      connection.go       GORM PostgreSQL setup
      migrations.go       AutoMigrate
      repositories/       IdempotencyRepo, PaymentRepo
    processor/
      simulator.go        Simulated payment processor
  presentation/echo/
    server.go             Echo setup, graceful shutdown
    routing.go            Route registration
    errorhandler.go       AppError to JSON mapping
    middleware/            TraceID, Recovery, Logger
    handlers/             PaymentHandler, HealthHandler
utils/
  config/                 .env loader with defaults
  fingerprint/            SHA256 request hashing
docs/                     Architecture, API, concurrency, infrastructure docs
tests/postman/            Postman collection and environment
tests/scripts/            Demo shell script
```
