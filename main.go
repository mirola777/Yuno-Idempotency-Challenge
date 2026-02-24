package main

import (
	"log"
	"os"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/application/use_cases"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/infrastructure/database"
	echoserver "github.com/mirola777/Yuno-Idempotency-Challenge/internal/presentation/echo"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/utils/config"
)

func main() {
	cfg := config.Load()

	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Printf("failed to connect to database: %v", err)
		os.Exit(1)
	}

	if err := database.RunMigrations(db); err != nil {
		log.Printf("failed to run migrations: %v", err)
		os.Exit(1)
	}

	container := use_cases.NewContainer(db, cfg)

	server := echoserver.NewServer(cfg)
	echoserver.ConfigureRoutes(server.Echo(), container)

	errC := server.Start()
	if err := <-errC; err != nil {
		log.Printf("server error: %v", err)
		os.Exit(1)
	}
}
