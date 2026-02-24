package gormdb

import (
	"context"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	"gorm.io/gorm"
)

type txKey struct{}

type TransactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) domain.TransactionManager {
	return &TransactionManager{db: db}
}

func (tm *TransactionManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}

func ExtractTx(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return fallback
}

func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}
