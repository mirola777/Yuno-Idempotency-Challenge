package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func clearEnvVars(t *testing.T) {
	t.Helper()
	vars := []string{
		"APP_PORT", "DB_HOST", "DB_PORT", "DB_USER",
		"DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"IDEMPOTENCY_KEY_TTL", "CLEANUP_INTERVAL", "GRACEFUL_TIMEOUT",
	}
	for _, v := range vars {
		os.Unsetenv(v)
	}
}

func TestLoad_DefaultValues(t *testing.T) {
	clearEnvVars(t)

	cfg := Load()

	assert.Equal(t, "8080", cfg.AppPort)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, "5432", cfg.DBPort)
	assert.Equal(t, "idempotency", cfg.DBUser)
	assert.Equal(t, "idempotency123", cfg.DBPassword)
	assert.Equal(t, "idempotency_db", cfg.DBName)
	assert.Equal(t, "disable", cfg.DBSSLMode)
	assert.Equal(t, 24*time.Hour, cfg.IdempotencyKeyTTL)
	assert.Equal(t, time.Hour, cfg.CleanupInterval)
	assert.Equal(t, 5*time.Second, cfg.GracefulTimeout)
}

func TestLoad_ReadsEnvVars(t *testing.T) {
	clearEnvVars(t)

	t.Setenv("APP_PORT", "9090")
	t.Setenv("DB_HOST", "db.example.com")
	t.Setenv("IDEMPOTENCY_KEY_TTL", "48h")

	cfg := Load()

	assert.Equal(t, "9090", cfg.AppPort)
	assert.Equal(t, "db.example.com", cfg.DBHost)
	assert.Equal(t, 48*time.Hour, cfg.IdempotencyKeyTTL)
}

func TestDSN(t *testing.T) {
	cfg := &Config{
		DBHost:     "localhost",
		DBUser:     "user",
		DBPassword: "pass",
		DBName:     "testdb",
		DBPort:     "5432",
		DBSSLMode:  "disable",
	}

	expected := "host=localhost user=user password=pass dbname=testdb port=5432 sslmode=disable"
	assert.Equal(t, expected, cfg.DSN())
}

func TestParseDuration_ValidDuration(t *testing.T) {
	d := parseDuration("30m", time.Hour)

	assert.Equal(t, 30*time.Minute, d)
}

func TestParseDuration_InvalidFallback(t *testing.T) {
	d := parseDuration("not-a-duration", 5*time.Second)

	assert.Equal(t, 5*time.Second, d)
}
