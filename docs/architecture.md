# Architecture

This project follows a **Hexagonal Architecture** (also known as Ports and Adapters), which enforces a strict separation between business logic and infrastructure concerns. The core idea is that the domain layer defines interfaces (ports) and all external systems connect through implementations of those interfaces (adapters).

---

## Layer Responsibilities

### domain/

Pure domain models, custom error types, and port interfaces. This layer has **zero external dependencies** -- it defines what the application does, not how.

- `models.go` -- Data structures: `Payment`, `PaymentRequest`, `IdempotencyRecord`, along with enums for `PaymentStatus`, `Currency`, and `IdempotencyStatus`.
- `errors/base.go` -- `AppError` struct with a `Messages` map for per-language translations, a `Localize(lang)` method that returns a localized copy, and a `newAppError()` constructor.
- `errors/payment.go` -- Error factory functions. Each factory embeds its own `Messages{"en": "...", "es": "..."}` map with translations.
- `ports.go` -- Interface definitions: `TransactionManager`, `IdempotencyRepository`, `PaymentRepository`, and `PaymentProcessor`. The domain layer has zero infrastructure imports -- transaction management is abstracted through the `TransactionManager` interface.

### application/

Use cases and service orchestration. Depends **only on domain** interfaces. Contains the core business rules for idempotent payment creation, validation, and the cleanup background loop.

- `use_cases/create_payment.go` -- Idempotency engine. Handles key validation, request fingerprint comparison, transactional payment creation, and cached response retrieval.
- `use_cases/get_payment.go` -- Payment retrieval by ID.
- `use_cases/get_by_idempotency_key.go` -- Idempotency key lookup.
- `use_cases/container.go` -- Dependency injection container. Handles DB connection and migrations internally, wires concrete infrastructure implementations to domain interfaces, and starts the background cleanup goroutine. `NewContainer` takes only `*config.Config` and returns `(*Container, error)`.

### infrastructure/

Secondary adapters that implement domain interfaces. This layer talks to external systems (database, payment processors) but is driven by the domain contracts.

- `gorm/connection.go` -- PostgreSQL connection setup via GORM with connection pool configuration (25 max open, 10 max idle, 5min lifetime). Package name is `gormdb`.
- `gorm/transaction.go` -- Implements `domain.TransactionManager`. Uses context-based transaction propagation: stores the active `*gorm.DB` transaction in the context via `context.WithValue`, and repositories extract it with `ExtractTx(ctx, fallback)`.
- `gorm/migrations.go` -- Migration runner that delegates to the `migrations/` subdirectory.
- `gorm/migrations/` -- Individual schema migration definitions for payments and idempotency records.
- `gorm/repositories/idempotency_repo.go` -- Implements `domain.IdempotencyRepository` with `SELECT ... FOR UPDATE` locking support.
- `gorm/repositories/payment_repo.go` -- Implements `domain.PaymentRepository`.
- `gorm/testdb.go` -- Test database helpers for repository tests.
- `processor/simulator.go` -- Implements `domain.PaymentProcessor`. Simulates payment processing with configurable outcomes based on test card numbers.

### presentation/

Primary adapters that expose the application to the outside world. Calls application services exclusively through domain interfaces.

- `echo/server.go` -- Echo HTTP server setup and graceful shutdown on SIGTERM/SIGINT/SIGQUIT. `NewServer` accepts the container directly and configures routes internally.
- `echo/routing.go` -- Route registration: maps HTTP endpoints to handler methods and applies middleware.
- `echo/handlers/payment_handler.go` -- HTTP handlers for payment creation, retrieval, and idempotency key lookup.
- `echo/handlers/health_handler.go` -- Health check endpoint.
- `echo/middleware/middleware.go` -- Cross-cutting middleware: trace ID propagation, request logging, and panic recovery.
- `echo/errorhandler.go` -- Custom error handler that translates `domain.AppError` into structured JSON responses. Uses `AppError.Localize(lang)` with the `Accept-Language` header for localized error messages.

### utils/

Cross-cutting utilities that do not belong to any specific architectural layer.

- `config/config.go` -- Environment-aware configuration loader. Reads from `.env` file with OS environment variable fallback. Supports three environments (`dev`, `test`, `prod`) via `APP_ENV`, with helper methods `IsDev()`, `IsProd()`, and `IsTest()` for environment-specific behavior. Parses duration values for TTL, cleanup interval, and graceful shutdown timeout.
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

- **presentation** calls application use cases directly via the `Container` struct.
- **application** calls `domain.TransactionManager`, `domain.IdempotencyRepository`, `domain.PaymentRepository`, and `domain.PaymentProcessor` (all implemented by infrastructure).
- **domain** defines all interfaces; it depends on nothing.
- **infrastructure** implements domain interfaces; it depends only on domain.

---

## Dependency Injection

The `container.go` file in `internal/application/use_cases/` serves as the composition root. `NewContainer` takes only `*config.Config` and returns `(*Container, error)`. It performs all wiring internally:

1. Opens the database connection via `gormdb.NewConnection()`.
2. Runs migrations via `gormdb.RunMigrations()`.
3. Creates the `TransactionManager` via `gormdb.NewTransactionManager()`.
4. Creates concrete repository implementations (`NewIdempotencyRepo`, `NewPaymentRepo`).
5. Creates the payment processor simulator (`NewSimulator`).
6. Creates the use case instances (`CreatePaymentUseCase`, `GetPaymentUseCase`, `GetByIdempotencyKeyUseCase`).
7. Starts the background cleanup goroutine.
8. Exposes the assembled use cases through the `Container` struct.

The `main.go` entry point is fully agnostic -- it has no database imports and only knows about config, container, and server:

1. Loads configuration via `config.Load()`.
2. Creates the dependency container via `use_cases.NewContainer(cfg)` (DB connection and migrations happen inside).
3. Creates the Echo server via `echoserver.NewServer(cfg, container)` (route configuration happens inside).
4. Starts the server and blocks until shutdown.

---

## Interface Contracts

All cross-layer communication happens through interfaces defined in `domain/ports.go`:

```go
type TransactionManager interface {
    RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type IdempotencyRepository interface {
    FindByKey(ctx context.Context, key string) (*IdempotencyRecord, error)
    FindByKeyForUpdate(ctx context.Context, key string) (*IdempotencyRecord, error)
    Create(ctx context.Context, record *IdempotencyRecord) error
    Update(ctx context.Context, record *IdempotencyRecord) error
    DeleteExpired(ctx context.Context) (int64, error)
}

type PaymentRepository interface {
    Create(ctx context.Context, payment *Payment) error
    FindByID(ctx context.Context, id string) (*Payment, error)
}

type PaymentProcessor interface {
    Process(ctx context.Context, req PaymentRequest) (*Payment, error)
}
```

Notice that the domain interfaces contain **no infrastructure types** (no `*gorm.DB`, no ORM-specific parameters). Transaction state is propagated through `context.Context` -- the `TransactionManager` injects a transaction handle into the context, and repositories extract it transparently. This keeps the domain layer completely pure while still supporting transactional operations.

This design allows any layer to be replaced independently -- swap PostgreSQL for DynamoDB, replace the simulator with a real payment gateway, or switch Echo for Gin -- without touching the domain or application logic.
