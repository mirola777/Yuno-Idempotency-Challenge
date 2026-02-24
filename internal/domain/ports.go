package domain

import "context"

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
