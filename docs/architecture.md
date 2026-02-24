# Architecture

This project follows a **Hexagonal Architecture** (also known as Ports and Adapters), which enforces a strict separation between business logic and infrastructure concerns. The core idea is that the domain layer defines interfaces (ports) and all external systems connect through implementations of those interfaces (adapters).

---

## Layer Responsibilities

### domain/

Pure domain models, custom error types, and port interfaces. This layer has **zero external dependencies** -- it defines what the application does, not how.

- `models.go` -- Data structures: `Payment`, `PaymentRequest`, `IdempotencyRecord`, along with enums for `PaymentStatus`, `Currency`, and `IdempotencyStatus`.
- `errors.go` -- Structured application errors (`AppError`) with HTTP status codes and machine-readable error codes.
- `ports.go` -- Interface definitions: `IdempotencyRepository`, `PaymentRepository`, `PaymentProcessor`, and `PaymentService`.

### application/

Use cases and service orchestration. Depends **only on domain** interfaces. Contains the core business rules for idempotent payment creation, validation, and the cleanup background loop.

- `payment/service.go` -- Implements `domain.PaymentService`. Handles idempotency key validation, request fingerprint comparison, transactional payment creation, and cached response retrieval.
- `payment/container.go` -- Dependency injection container. Wires concrete infrastructure implementations to domain interfaces and starts the background cleanup goroutine.

### infrastructure/

Secondary adapters that implement domain interfaces. This layer talks to external systems (database, payment processors) but is driven by the domain contracts.

- `database/connection.go` -- PostgreSQL connection setup via GORM with connection pool configuration (25 max open, 10 max idle, 5min lifetime).
- `database/migrations.go` -- GORM `AutoMigrate` for `Payment` and `IdempotencyRecord` tables.
- `database/repositories/idempotency_repo.go` -- Implements `domain.IdempotencyRepository` with `SELECT ... FOR UPDATE` locking support.
- `database/repositories/payment_repo.go` -- Implements `domain.PaymentRepository`.
- `processor/simulator.go` -- Implements `domain.PaymentProcessor`. Simulates payment processing with configurable outcomes based on test card numbers.

### presentation/

Primary adapters that expose the application to the outside world. Calls application services exclusively through domain interfaces.

- `echo/server.go` -- Echo HTTP server with graceful shutdown on SIGTERM/SIGINT/SIGQUIT.
- `echo/routing.go` -- Route registration: maps HTTP endpoints to handler methods and applies middleware.
- `echo/handlers/payment_handler.go` -- HTTP handlers for payment creation, retrieval, and idempotency key lookup.
- `echo/handlers/health_handler.go` -- Health check endpoint.
- `echo/middleware/middleware.go` -- Cross-cutting middleware: trace ID propagation, request logging, and panic recovery.
- `echo/errorhandler.go` -- Custom error handler that translates `domain.AppError` into structured JSON responses.

### utils/

Cross-cutting utilities that do not belong to any specific architectural layer.

- `config/config.go` -- Configuration loader. Reads from `.env` file with OS environment variable fallback. Parses duration values for TTL, cleanup interval, and graceful shutdown timeout.
- `fingerprint/fingerprint.go` -- Computes a SHA-256 hash of the payment request body. Used to detect payload mismatches on idempotency key reuse.

---

## Dependency Flow

All dependencies flow inward toward the domain. The domain layer depends on nothing. Infrastructure and presentation layers depend on the domain, never on each other.

```
  +----------------+       +---------------+       +--------+       +------------------+
  |  presentation  | ----> |  application  | ----> | domain | <---- | infrastructure   |
  |  (Echo HTTP)   |       |  (services)   |       | (core) |       | (repos, processor)|
  +----------------+       +---------------+       +--------+       +------------------+
```

In dependency terms:

```
presentation --> application --> domain <-- infrastructure
```

- **presentation** calls `domain.PaymentService` (implemented by application).
- **application** calls `domain.IdempotencyRepository`, `domain.PaymentRepository`, and `domain.PaymentProcessor` (implemented by infrastructure).
- **domain** defines all interfaces; it depends on nothing.
- **infrastructure** implements domain interfaces; it depends only on domain.

---

## Dependency Injection

The `container.go` file in `internal/application/payment/` serves as the composition root. It performs all wiring:

1. Creates concrete repository implementations (`NewIdempotencyRepo`, `NewPaymentRepo`).
2. Creates the payment processor simulator (`NewSimulator`).
3. Passes all implementations into `NewService`, which returns a `domain.PaymentService`.
4. Starts the background cleanup goroutine.
5. Exposes the assembled `PaymentService` through the `Container` struct.

The `main.go` entry point orchestrates startup:

1. Loads configuration via `config.Load()`.
2. Opens the database connection via `database.NewConnection()`.
3. Runs auto-migrations via `database.RunMigrations()`.
4. Creates the dependency container via `payment.NewContainer()`.
5. Creates the Echo server and configures routes, injecting `container.PaymentService`.
6. Starts the server and blocks until shutdown.

---

## Interface Contracts

All cross-layer communication happens through interfaces defined in `domain/ports.go`:

```go
type IdempotencyRepository interface {
    FindByKey(ctx context.Context, key string) (*IdempotencyRecord, error)
    Create(ctx context.Context, record *IdempotencyRecord) error
    Update(ctx context.Context, record *IdempotencyRecord) error
    DeleteExpired(ctx context.Context) (int64, error)
    FindByKeyForUpdate(ctx context.Context, tx *gorm.DB, key string) (*IdempotencyRecord, error)
    CreateInTx(ctx context.Context, tx *gorm.DB, record *IdempotencyRecord) error
    UpdateInTx(ctx context.Context, tx *gorm.DB, record *IdempotencyRecord) error
}

type PaymentRepository interface {
    Create(ctx context.Context, payment *Payment) error
    FindByID(ctx context.Context, id string) (*Payment, error)
    CreateInTx(ctx context.Context, tx *gorm.DB, payment *Payment) error
}

type PaymentProcessor interface {
    Process(ctx context.Context, req PaymentRequest) (*Payment, error)
}

type PaymentService interface {
    CreatePayment(ctx context.Context, idempotencyKey string, req PaymentRequest) (*Payment, error)
    GetPayment(ctx context.Context, paymentID string) (*Payment, error)
    GetByIdempotencyKey(ctx context.Context, key string) (*IdempotencyRecord, error)
}
```

This design allows any layer to be replaced independently -- swap PostgreSQL for DynamoDB, replace the simulator with a real payment gateway, or switch Echo for Gin -- without touching the domain or application logic.
