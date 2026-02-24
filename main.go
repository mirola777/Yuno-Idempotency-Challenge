package main

import (
	"log"
	"os"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/application/use_cases"
	echoserver "github.com/mirola777/Yuno-Idempotency-Challenge/internal/presentation/echo"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/utils/config"
)

func main() {
	cfg := config.Load()

	container, err := use_cases.NewContainer(cfg)
	if err != nil {
		log.Printf("failed to initialize application: %v", err)
		os.Exit(1)
	}

	server := echoserver.NewServer(cfg, container)

	if err := <-server.Start(); err != nil {
		log.Printf("server error: %v", err)
		os.Exit(1)
	}
}
