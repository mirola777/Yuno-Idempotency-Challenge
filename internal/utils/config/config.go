package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Environment string

const (
	EnvTest        Environment = "test"
	EnvDevelopment Environment = "dev"
	EnvProduction  Environment = "prod"
)

type Config struct {
	AppEnv           Environment
	AppPort          string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBSSLMode        string
	IdempotencyKeyTTL time.Duration
	CleanupInterval  time.Duration
	GracefulTimeout  time.Duration
}

func (c *Config) IsDev() bool {
	return c.AppEnv == EnvDevelopment
}

func (c *Config) IsProd() bool {
	return c.AppEnv == EnvProduction
}

func (c *Config) IsTest() bool {
	return c.AppEnv == EnvTest
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		AppEnv:           parseEnv(getEnv("APP_ENV", "dev")),
		AppPort:          getEnv("APP_PORT", "8080"),
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USER", "idempotency"),
		DBPassword:       getEnv("DB_PASSWORD", "idempotency123"),
		DBName:           getEnv("DB_NAME", "idempotency_db"),
		DBSSLMode:        getEnv("DB_SSLMODE", "disable"),
		IdempotencyKeyTTL: parseDuration(getEnv("IDEMPOTENCY_KEY_TTL", "24h"), 24*time.Hour),
		CleanupInterval:  parseDuration(getEnv("CLEANUP_INTERVAL", "1h"), time.Hour),
		GracefulTimeout:  parseDuration(getEnv("GRACEFUL_TIMEOUT", "5s"), 5*time.Second),
	}
}

func (c *Config) DSN() string {
	return "host=" + c.DBHost +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" port=" + c.DBPort +
		" sslmode=" + c.DBSSLMode
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func parseDuration(value string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return d
}

func parseEnv(value string) Environment {
	switch value {
	case "prod", "production":
		return EnvProduction
	case "test":
		return EnvTest
	default:
		return EnvDevelopment
	}
}
