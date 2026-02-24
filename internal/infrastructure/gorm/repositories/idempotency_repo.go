package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	gormdb "github.com/mirola777/Yuno-Idempotency-Challenge/internal/infrastructure/gorm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IdempotencyRepo struct {
	db *gorm.DB
}

func NewIdempotencyRepo(db *gorm.DB) domain.IdempotencyRepository {
	return &IdempotencyRepo{db: db}
}

func (r *IdempotencyRepo) conn(ctx context.Context) *gorm.DB {
	return gormdb.ExtractTx(ctx, r.db).WithContext(ctx)
}

func (r *IdempotencyRepo) FindByKey(ctx context.Context, key string) (*domain.IdempotencyRecord, error) {
	var record domain.IdempotencyRecord
	err := r.conn(ctx).
		Where("key = ? AND expires_at > ?", key, time.Now()).
		First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *IdempotencyRepo) FindByKeyForUpdate(ctx context.Context, key string) (*domain.IdempotencyRecord, error) {
	var record domain.IdempotencyRecord
	err := r.conn(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("key = ? AND expires_at > ?", key, time.Now()).
		First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *IdempotencyRepo) Create(ctx context.Context, record *domain.IdempotencyRecord) error {
	return r.conn(ctx).Create(record).Error
}

func (r *IdempotencyRepo) Update(ctx context.Context, record *domain.IdempotencyRecord) error {
	return r.conn(ctx).Save(record).Error
}

func (r *IdempotencyRepo) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.conn(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&domain.IdempotencyRecord{})
	return result.RowsAffected, result.Error
}
