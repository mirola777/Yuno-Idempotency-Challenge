# Infrastructure

## Environments

The application supports three environments controlled by the `APP_ENV` variable:

| Environment | Value | Description |
|---|---|---|
| Development | `dev` | Default mode. Used for local development. |
| Test | `test` | Used when running the test suite. |
| Production | `prod` | Used for deployed environments with optimized settings. |

The `Config` struct provides helper methods for environment checks:

```go
func (c *Config) IsDev() bool  { return c.AppEnv == EnvDevelopment }
func (c *Config) IsProd() bool { return c.AppEnv == EnvProduction }
func (c *Config) IsTest() bool { return c.AppEnv == EnvTest }
```

These methods allow any layer of the application to branch behavior based on the current environment without string comparisons. The `parseEnv` function normalizes the input, accepting both `"prod"` and `"production"` for the production environment, while any unrecognized value defaults to development.

---

## Docker

### Multi-Stage Dockerfile

The Dockerfile uses a multi-stage build with four stages:

```dockerfile
FROM golang:1.24-alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM base AS dev
RUN go install github.com/air-verse/air@latest
CMD ["air", "-c", ".air.toml"]

FROM base AS build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server .

FROM alpine:3.20 AS prod
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build /app/server .
EXPOSE 8080
CMD ["./server"]
```

| Stage | Purpose |
|---|---|
| `base` | Shared foundation. Downloads Go modules and copies source code. |
| `dev` | Extends `base`. Installs Air for hot reload. Used with volume mounts so source changes trigger automatic rebuilds. |
| `build` | Extends `base`. Compiles a statically linked, stripped binary (`-ldflags="-s -w"`, `CGO_ENABLED=0`). This stage is not run directly; its output is consumed by the `prod` stage. |
| `prod` | Starts from a minimal `alpine:3.20` image. Copies only the compiled binary from the `build` stage. The final image contains no Go toolchain, no source code, and no build artifacts. |

### Docker Compose Profiles

The `docker-compose.yml` defines three services using Compose profiles to separate development and production workflows:

**PostgreSQL (always runs):**

The `postgres` service runs regardless of the selected profile. It provides the PostgreSQL 16 database with a health check, named volume for persistence, and environment variable configuration.

**Dev profile (`--profile dev`):**

```bash
docker compose --profile dev up --build
```

- Targets the `dev` stage of the Dockerfile
- Volume-mounts the project root (`.:/app`) so code changes are reflected immediately
- Air watches for `.go` and `.toml` file changes and rebuilds automatically
- Sets `APP_ENV=dev` and `DB_HOST=postgres`
- Depends on the PostgreSQL health check before starting

**Prod profile (`--profile prod`):**

```bash
docker compose --profile prod up --build
```

- Targets the `prod` stage of the Dockerfile
- No volume mounts; the container runs the compiled binary only
- Sets `APP_ENV=prod` and `DB_HOST=postgres`
- Configured with `restart: unless-stopped` for resilience
- Depends on the PostgreSQL health check before starting

**Database only (no profile):**

```bash
docker compose up -d
```

Starts only PostgreSQL. Useful when running the application locally with `go run` or `air`.

---

## Air Hot Reload

The project includes an `.air.toml` configuration file for [Air](https://github.com/air-verse/air), a live-reloading tool for Go applications.

### Configuration

| Setting | Value | Description |
|---|---|---|
| Build command | `go build -o ./tmp/main .` | Compiles the binary into the `tmp` directory |
| Rebuild delay | 1000ms | Waits 1 second after a file change before rebuilding |
| Watched extensions | `.go`, `.toml` | Only Go source files and TOML config files trigger rebuilds |
| Excluded directories | `tmp`, `vendor`, `tests`, `docs` | Prevents unnecessary rebuilds |
| Excluded patterns | `_test\.go$` | Test files do not trigger rebuilds |
| Clean on exit | `true` | Removes the `tmp` directory when Air stops |

### Usage

**Locally:**

```bash
air
```

Requires Air to be installed: `go install github.com/air-verse/air@latest`

**In Docker:**

```bash
docker compose --profile dev up --build
```

Air is installed automatically in the `dev` Dockerfile stage. The volume mount ensures local file changes are visible inside the container.

---

## Environment Variables

Configuration is loaded from a `.env` file at the project root, with OS environment variables taking precedence. If neither is set, the default value is used.

| Variable | Description | Default |
|---|---|---|
| `APP_ENV` | Application environment (`dev`, `test`, `prod`) | `dev` |
| `APP_PORT` | HTTP server listening port | `8080` |
| `DB_HOST` | PostgreSQL hostname | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL username | `idempotency` |
| `DB_PASSWORD` | PostgreSQL password | `idempotency123` |
| `DB_NAME` | PostgreSQL database name | `idempotency_db` |
| `DB_SSLMODE` | PostgreSQL SSL mode (`disable`, `require`, etc.) | `disable` |
| `IDEMPOTENCY_KEY_TTL` | How long idempotency keys remain valid (Go duration) | `24h` |
| `CLEANUP_INTERVAL` | Interval between expired record cleanup runs | `1h` |
| `GRACEFUL_TIMEOUT` | Maximum time to wait for in-flight requests on shutdown | `5s` |

Duration values use Go's `time.ParseDuration` format: `5s`, `1m`, `24h`, `500ms`, etc.

---

## Running the Application

### Development with Docker (recommended)

1. Copy the example environment file:

   ```bash
   cp .env.example .env
   ```

2. Start everything with the dev profile:

   ```bash
   docker compose --profile dev up --build
   ```

   This starts PostgreSQL and the application with Air hot reload. Edit any `.go` file and the server rebuilds automatically.

### Development without Docker

1. Start only the database:

   ```bash
   docker compose up -d
   ```

2. Run with Air:

   ```bash
   air
   ```

   Or without Air:

   ```bash
   go run main.go
   ```

### Production with Docker

```bash
docker compose --profile prod up --build -d
```

This builds the optimized binary and runs it in a minimal container with automatic restart.

### Without Docker (existing PostgreSQL)

If you already have a PostgreSQL instance, configure the connection by setting environment variables or editing the `.env` file:

```bash
export APP_ENV=prod
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

The `internal/utils/config/config.go` module handles configuration with a two-step fallback:

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
CLEANUP_INTERVAL=30m
CLEANUP_INTERVAL=6h
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
GRACEFUL_TIMEOUT=10s
GRACEFUL_TIMEOUT=30s
```

In production, set this to a value that accommodates your slowest expected request (payment processing takes 50-200ms in the simulator, but a real processor may take longer).
