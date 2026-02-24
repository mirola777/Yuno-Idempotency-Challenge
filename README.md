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
- Air (optional, for local hot reload in development)

### Setup

```bash
cp .env.example .env
```

The application supports three environment modes controlled by the `APP_ENV` variable: `dev` (default), `test`, and `prod`. See the [Environments](#environments) section for details.

**Development with Docker (hot reload):**

```bash
docker compose --profile dev up --build
```

This starts PostgreSQL and the application with Air hot reload. Source code is volume-mounted, so changes are picked up automatically.

**Production with Docker:**

```bash
docker compose --profile prod up --build
```

This builds an optimized binary and runs it in a minimal Alpine container.

**Local development (database only in Docker):**

```bash
docker compose up -d
air
```

Or without Air:

```bash
docker compose up -d
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

## Environments

The `APP_ENV` variable controls the application environment. It accepts three values:

| Value | Description |
|---|---|
| `dev` | Development mode (default). Used for local development with verbose output. |
| `test` | Test mode. Used when running the test suite. |
| `prod` | Production mode. Used for deployed environments with optimized settings. |

The configuration provides helper methods `IsDev()`, `IsProd()`, and `IsTest()` for environment-specific branching in application code.

### Running in Each Environment

**Development:**

```bash
APP_ENV=dev go run main.go
```

Or with Air for hot reload:

```bash
APP_ENV=dev air
```

Or fully containerized:

```bash
docker compose --profile dev up --build
```

**Production:**

```bash
APP_ENV=prod go run main.go
```

Or containerized with the optimized binary:

```bash
docker compose --profile prod up --build
```

**Tests:**

```bash
APP_ENV=test go test ./...
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
   - Duplicate request (same key + same payload): returns cached response with `X-Idempotent-Replayed: true` header
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
APP_ENV=test go test ./...
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
| APP_ENV | dev | Application environment: `dev`, `test`, or `prod` |
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

## Docker

### Multi-Stage Dockerfile

The Dockerfile uses a multi-stage build with three stages:

| Stage | Purpose |
|---|---|
| `base` | Downloads Go dependencies and copies source code |
| `dev` | Installs Air and runs with hot reload |
| `prod` | Builds an optimized static binary and runs in a minimal Alpine image |

### Docker Compose Profiles

The `docker-compose.yml` defines two application profiles alongside a shared PostgreSQL service:

**Dev profile** (`docker compose --profile dev up`):
- Builds from the `dev` Dockerfile stage
- Mounts the project directory as a volume for live code changes
- Runs Air for automatic rebuilds on file changes
- Sets `APP_ENV=dev`

**Prod profile** (`docker compose --profile prod up`):
- Builds from the `prod` Dockerfile stage
- Produces a small container with only the compiled binary
- Sets `APP_ENV=prod` and `restart: unless-stopped`

Running `docker compose up -d` without a profile starts only the PostgreSQL database, which is useful for local development with `go run` or `air`.

## Project Structure

```
.air.toml                 Air hot reload configuration
Dockerfile                Multi-stage build (dev with Air, prod optimized)
docker-compose.yml        PostgreSQL + app profiles (dev, prod)
main.go                   Entry point (config, container, server -- no DB imports)
internal/
  application/use_cases/
    container.go          DI wiring, DB connection, migrations
    create_payment.go     Idempotency engine
    get_payment.go        Payment retrieval
    get_by_idempotency_key.go  Key lookup
    create_payment_test.go     Unit tests
  domain/
    models.go             Payment, IdempotencyRecord, enums
    errors/
      base.go             AppError with Messages map and Localize(lang)
      payment.go          Error factories with embedded translations
    ports.go              TransactionManager, Repository, Processor interfaces
  infrastructure/
    gorm/
      connection.go       GORM PostgreSQL setup (package gormdb)
      transaction.go      TransactionManager with context-based tx propagation
      migrations.go       Migration runner
      migrations/         Schema migration definitions
      repositories/       IdempotencyRepo, PaymentRepo
      testdb.go           Test database helpers
    processor/
      simulator.go        Simulated payment processor
  presentation/echo/
    server.go             Echo setup, route wiring, graceful shutdown
    routing.go            Route registration
    errorhandler.go       AppError to JSON mapping with localization
    middleware/            TraceID, Recovery, Logger
    handlers/             PaymentHandler, HealthHandler
  utils/
    config/               Environment-aware config with .env loader (APP_ENV support)
    fingerprint/          SHA256 request hashing
docs/                     Architecture, API, concurrency, infrastructure docs
tests/postman/            Postman collection and environment
tests/scripts/            Demo shell script
tests/integration/        Integration tests
```
