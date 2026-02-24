package repositories

import (
	"context"
	"errors"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	gormdb "github.com/mirola777/Yuno-Idempotency-Challenge/internal/infrastructure/gorm"
	"gorm.io/gorm"
)

type PaymentRepo struct {
	db *gorm.DB
}

func NewPaymentRepo(db *gorm.DB) domain.PaymentRepository {
	return &PaymentRepo{db: db}
}

func (r *PaymentRepo) conn(ctx context.Context) *gorm.DB {
	return gormdb.ExtractTx(ctx, r.db).WithContext(ctx)
}

func (r *PaymentRepo) Create(ctx context.Context, payment *domain.Payment) error {
	return r.conn(ctx).Create(payment).Error
}

func (r *PaymentRepo) FindByID(ctx context.Context, id string) (*domain.Payment, error) {
	var payment domain.Payment
	err := r.conn(ctx).Where("id = ?", id).First(&payment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &payment, nil
}
