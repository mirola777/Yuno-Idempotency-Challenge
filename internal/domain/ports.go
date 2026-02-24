package domain

import (
	"context"

	"gorm.io/gorm"
)

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
