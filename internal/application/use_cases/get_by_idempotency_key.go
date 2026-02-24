package use_cases

import (
	"context"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	apperrors "github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain/errors"
)

type GetByIdempotencyKeyUseCase struct {
	idempotencyRepo domain.IdempotencyRepository
}

func NewGetByIdempotencyKeyUseCase(idempotencyRepo domain.IdempotencyRepository) *GetByIdempotencyKeyUseCase {
	return &GetByIdempotencyKeyUseCase{
		idempotencyRepo: idempotencyRepo,
	}
}

func (uc *GetByIdempotencyKeyUseCase) Execute(ctx context.Context, key string) (*domain.IdempotencyRecord, error) {
	record, err := uc.idempotencyRepo.FindByKey(ctx, key)
	if err != nil {
		return nil, apperrors.ErrInternal()
	}
	if record == nil {
		return nil, apperrors.ErrIdempotencyKeyNotFound()
	}
	return record, nil
}
