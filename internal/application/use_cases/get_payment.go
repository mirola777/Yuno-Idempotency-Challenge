package use_cases

import (
	"context"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	apperrors "github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain/errors"
)

type GetPaymentUseCase struct {
	paymentRepo domain.PaymentRepository
}

func NewGetPaymentUseCase(paymentRepo domain.PaymentRepository) *GetPaymentUseCase {
	return &GetPaymentUseCase{
		paymentRepo: paymentRepo,
	}
}

func (uc *GetPaymentUseCase) Execute(ctx context.Context, paymentID string) (*domain.Payment, error) {
	payment, err := uc.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, apperrors.ErrInternal()
	}
	if payment == nil {
		return nil, apperrors.ErrPaymentNotFound()
	}
	return payment, nil
}
