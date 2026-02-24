package use_cases

import (
	"context"
	"log"
	"time"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	gormdb "github.com/mirola777/Yuno-Idempotency-Challenge/internal/infrastructure/gorm"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/infrastructure/gorm/repositories"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/infrastructure/processor"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/utils/config"
)

type Container struct {
	CreatePayment       *CreatePaymentUseCase
	GetPayment          *GetPaymentUseCase
	GetByIdempotencyKey *GetByIdempotencyKeyUseCase
}

func NewContainer(cfg *config.Config) (*Container, error) {
	db, err := gormdb.NewConnection(cfg)
	if err != nil {
		return nil, err
	}

	if err := gormdb.RunMigrations(db); err != nil {
		return nil, err
	}

	idempotencyRepo := repositories.NewIdempotencyRepo(db)
	paymentRepo := repositories.NewPaymentRepo(db)
	paymentProcessor := processor.NewSimulator()

	txManager := gormdb.NewTransactionManager(db)

	createPayment := NewCreatePaymentUseCase(txManager, idempotencyRepo, paymentRepo, paymentProcessor, cfg.IdempotencyKeyTTL)
	getPayment := NewGetPaymentUseCase(paymentRepo)
	getByIdempotencyKey := NewGetByIdempotencyKeyUseCase(idempotencyRepo)

	go startCleanupLoop(idempotencyRepo, cfg.CleanupInterval)

	return &Container{
		CreatePayment:       createPayment,
		GetPayment:          getPayment,
		GetByIdempotencyKey: getByIdempotencyKey,
	}, nil
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
