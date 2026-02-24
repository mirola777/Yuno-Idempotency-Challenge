package payment

import (
	"context"
	"log"
	"time"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/infrastructure/database/repositories"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/infrastructure/processor"
	"github.com/mirola777/Yuno-Idempotency-Challenge/utils/config"
	"gorm.io/gorm"
)

type Container struct {
	PaymentService domain.PaymentService
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	idempotencyRepo := repositories.NewIdempotencyRepo(db)
	paymentRepo := repositories.NewPaymentRepo(db)
	paymentProcessor := processor.NewSimulator()

	paymentService := NewService(
		db,
		idempotencyRepo,
		paymentRepo,
		paymentProcessor,
		cfg.IdempotencyKeyTTL,
	)

	go startCleanupLoop(idempotencyRepo, cfg.CleanupInterval)

	return &Container{
		PaymentService: paymentService,
	}
}

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
