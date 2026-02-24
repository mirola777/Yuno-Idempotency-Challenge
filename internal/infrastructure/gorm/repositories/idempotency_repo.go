package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IdempotencyRepo struct {
	db *gorm.DB
}

func NewIdempotencyRepo(db *gorm.DB) domain.IdempotencyRepository {
	return &IdempotencyRepo{db: db}
}

func (r *IdempotencyRepo) FindByKey(ctx context.Context, key string) (*domain.IdempotencyRecord, error) {
	var record domain.IdempotencyRecord
	err := r.db.WithContext(ctx).
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
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *IdempotencyRepo) Update(ctx context.Context, record *domain.IdempotencyRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

func (r *IdempotencyRepo) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&domain.IdempotencyRecord{})
	return result.RowsAffected, result.Error
}

func (r *IdempotencyRepo) FindByKeyForUpdate(ctx context.Context, tx *gorm.DB, key string) (*domain.IdempotencyRecord, error) {
	var record domain.IdempotencyRecord
	err := tx.WithContext(ctx).
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

func (r *IdempotencyRepo) CreateInTx(ctx context.Context, tx *gorm.DB, record *domain.IdempotencyRecord) error {
	return tx.WithContext(ctx).Create(record).Error
}

func (r *IdempotencyRepo) UpdateInTx(ctx context.Context, tx *gorm.DB, record *domain.IdempotencyRecord) error {
	return tx.WithContext(ctx).Save(record).Error
}
