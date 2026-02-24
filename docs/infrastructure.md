# Infrastructure

## Docker Compose Setup

The project includes a `docker-compose.yml` that provisions a **PostgreSQL 16 Alpine** instance. This is the only external dependency required to run the application.

```yaml
version: "3.9"

services:
  postgres:
    image: postgres:16-alpine
    container_name: idempotency-postgres
    environment:
      POSTGRES_USER: ${DB_USER:-idempotency}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-idempotency123}
      POSTGRES_DB: ${DB_NAME:-idempotency_db}
    ports:
      - "${DB_PORT:-5432}:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-idempotency} -d ${DB_NAME:-idempotency_db}"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  pgdata:
```

### Starting the Database

```bash
docker-compose up -d
```

This starts PostgreSQL on port 5432 (configurable via `DB_PORT`). The container uses a named volume (`pgdata`) so data persists across restarts.

To verify the database is ready:

```bash
docker-compose ps
```

The `healthcheck` configuration ensures the container reports healthy only after PostgreSQL accepts connections.

### Stopping the Database

```bash
docker-compose down
```

To also remove the persisted data volume:

```bash
docker-compose down -v
```

---

## Environment Variables

Configuration is loaded from a `.env` file at the project root, with OS environment variables taking precedence. If neither is set, the default value is used.

| Variable              | Description                                           | Default           |
|-----------------------|-------------------------------------------------------|-------------------|
| `APP_PORT`            | HTTP server listening port                            | `8080`            |
| `DB_HOST`             | PostgreSQL hostname                                   | `localhost`       |
| `DB_PORT`             | PostgreSQL port                                       | `5432`            |
| `DB_USER`             | PostgreSQL username                                   | `idempotency`     |
| `DB_PASSWORD`         | PostgreSQL password                                   | `idempotency123`  |
| `DB_NAME`             | PostgreSQL database name                              | `idempotency_db`  |
| `DB_SSLMODE`          | PostgreSQL SSL mode (`disable`, `require`, etc.)      | `disable`         |
| `IDEMPOTENCY_KEY_TTL` | How long idempotency keys remain valid (Go duration)  | `24h`             |
| `CLEANUP_INTERVAL`    | Interval between expired record cleanup runs          | `1h`              |
| `GRACEFUL_TIMEOUT`    | Maximum time to wait for in-flight requests on shutdown | `5s`            |

Duration values use Go's `time.ParseDuration` format: `5s`, `1m`, `24h`, `500ms`, etc.

---

## Running the Application

### With Docker Compose (recommended)

1. Copy the example environment file:

   ```bash
   cp .env.example .env
   ```

2. Start the PostgreSQL database:

   ```bash
   docker-compose up -d
   ```

3. Run the application:

   ```bash
   go run main.go
   ```

   The server starts on the configured `APP_PORT` (default: 8080).

### Without Docker (existing PostgreSQL)

If you already have a PostgreSQL instance, configure the connection by setting environment variables or editing the `.env` file:

```bash
export DB_HOST=your-postgres-host
export DB_PORT=5432
export DB_USER=your_user
export DB_PASSWORD=your_password
export DB_NAME=your_database
export DB_SSLMODE=require
```

Then run:

```bash
go run main.go
```

The application will connect to the specified PostgreSQL instance and auto-migrate the schema.

---

## Configuration Loading

The `utils/config/config.go` module handles configuration with a two-step fallback:

1. Attempt to load variables from a `.env` file in the working directory (using `godotenv`).
2. For each variable, check if an OS environment variable is set. If so, it takes precedence over the `.env` value.
3. If neither source provides a value, use the hardcoded default.

This means you can:
- Use a `.env` file for local development.
- Use OS environment variables in production (Docker, Kubernetes, systemd).
- Mix both -- OS env vars override `.env` values.

---

## Database

### Schema Management

The application uses **GORM AutoMigrate** to create and update database tables on startup. No manual migration scripts are needed.

```go
func RunMigrations(db *gorm.DB) error {
    return db.AutoMigrate(
        &domain.Payment{},
        &domain.IdempotencyRecord{},
    )
}
```

This creates two tables:

**payments:**

| Column       | Type         | Constraints       |
|--------------|--------------|-------------------|
| `id`         | varchar(36)  | PRIMARY KEY       |
| `amount`     | float        | NOT NULL          |
| `currency`   | varchar(3)   | NOT NULL          |
| `customer_id`| varchar(100) | NOT NULL          |
| `ride_id`    | varchar(100) | NOT NULL          |
| `status`     | varchar(20)  | NOT NULL          |
| `card_last_4`| varchar(4)   |                   |
| `description`| text         |                   |
| `fail_reason`| text         |                   |
| `created_at` | timestamp    | auto-generated    |

**idempotency_records:**

| Column               | Type         | Constraints       |
|----------------------|--------------|-------------------|
| `key`                | varchar(64)  | PRIMARY KEY       |
| `request_fingerprint`| varchar(64)  | NOT NULL          |
| `payment_id`         | varchar(36)  |                   |
| `response_body`      | jsonb        |                   |
| `status`             | varchar(20)  | NOT NULL          |
| `created_at`         | timestamp    | auto-generated    |
| `expires_at`         | timestamp    | NOT NULL, INDEXED |

### Connection Pool

The database connection is configured with the following pool settings in `database/connection.go`:

| Setting           | Value    | Description                                          |
|-------------------|----------|------------------------------------------------------|
| Max Open Conns    | 25       | Maximum number of open connections to the database   |
| Max Idle Conns    | 10       | Maximum number of idle connections in the pool       |
| Conn Max Lifetime | 5 min    | Maximum time a connection can be reused              |

These values are hardcoded for simplicity but can be extracted into configuration if needed.

---

## Background Cleanup

Expired idempotency records are cleaned up by a background goroutine that runs on a configurable interval (default: 1 hour).

```go
func startCleanupLoop(repo domain.IdempotencyRepository, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    for range ticker.C {
        cleaned, err := repo.DeleteExpired(context.Background())
        if err != nil {
            log.Printf("cleanup error: %v", err)
            continue
        }
        if cleaned > 0 {
            log.Printf("cleaned %d expired idempotency records", cleaned)
        }
    }
}
```

The cleanup deletes all records where `expires_at < NOW()`. The `expires_at` column has a database index to make this query efficient.

To adjust the cleanup frequency, set the `CLEANUP_INTERVAL` environment variable:

```bash
CLEANUP_INTERVAL=30m  # Run every 30 minutes
CLEANUP_INTERVAL=6h   # Run every 6 hours
```

---

## Graceful Shutdown

The server listens for `SIGTERM`, `SIGINT`, and `SIGQUIT` signals. When one is received:

1. The signal handler logs the shutdown initiation.
2. A context with the `GRACEFUL_TIMEOUT` deadline is created (default: 5 seconds).
3. `echo.Shutdown(ctx)` is called, which:
   - Stops accepting new connections.
   - Waits for in-flight requests to complete.
   - Returns an error if the timeout is exceeded.
4. The process exits.

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), s.config.GracefulTimeout)
defer cancel()

if err := s.echo.Shutdown(ctx); err != nil {
    errC <- err
}
```

To change the drain timeout:

```bash
GRACEFUL_TIMEOUT=10s   # Wait up to 10 seconds for requests to finish
GRACEFUL_TIMEOUT=30s   # Wait up to 30 seconds
```

In production, set this to a value that accommodates your slowest expected request (payment processing takes 50-200ms in the simulator, but a real processor may take longer).
