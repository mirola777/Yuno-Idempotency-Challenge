package use_cases

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	apperrors "github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain/errors"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/utils/fingerprint"
)

type CreatePaymentResult struct {
	Payment  *domain.Payment
	Replayed bool
}

type CreatePaymentUseCase struct {
	txManager       domain.TransactionManager
	idempotencyRepo domain.IdempotencyRepository
	paymentRepo     domain.PaymentRepository
	processor       domain.PaymentProcessor
	keyTTL          time.Duration
}

func NewCreatePaymentUseCase(
	txManager domain.TransactionManager,
	idempotencyRepo domain.IdempotencyRepository,
	paymentRepo domain.PaymentRepository,
	processor domain.PaymentProcessor,
	keyTTL time.Duration,
) *CreatePaymentUseCase {
	return &CreatePaymentUseCase{
		txManager:       txManager,
		idempotencyRepo: idempotencyRepo,
		paymentRepo:     paymentRepo,
		processor:       processor,
		keyTTL:          keyTTL,
	}
}

func (uc *CreatePaymentUseCase) Execute(ctx context.Context, idempotencyKey string, req domain.PaymentRequest) (*CreatePaymentResult, error) {
	if err := validateIdempotencyKey(idempotencyKey); err != nil {
		return nil, err
	}

	if err := validatePaymentRequest(req); err != nil {
		return nil, err
	}

	fp := fingerprint.Compute(req)

	var result *CreatePaymentResult
	var returnErr error

	txErr := uc.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		record, err := uc.idempotencyRepo.FindByKeyForUpdate(txCtx, idempotencyKey)
		if err != nil {
			returnErr = apperrors.ErrInternal()
			return err
		}

		if record != nil {
			if record.Status == domain.IdempotencyStatusProcessing {
				returnErr = apperrors.ErrPaymentProcessing()
				return returnErr
			}

			if record.RequestFingerprint != fp {
				returnErr = apperrors.ErrIdempotencyKeyConflict()
				return returnErr
			}

			var cached domain.Payment
			if err := json.Unmarshal(record.ResponseBody, &cached); err != nil {
				returnErr = apperrors.ErrInternal()
				return err
			}
			result = &CreatePaymentResult{Payment: &cached, Replayed: true}
			return nil
		}

		newRecord := &domain.IdempotencyRecord{
			Key:                idempotencyKey,
			RequestFingerprint: fp,
			Status:             domain.IdempotencyStatusProcessing,
			CreatedAt:          time.Now(),
			ExpiresAt:          time.Now().Add(uc.keyTTL),
		}
		if err := uc.idempotencyRepo.Create(txCtx, newRecord); err != nil {
			returnErr = apperrors.ErrInternal()
			return err
		}

		payment, err := uc.processor.Process(ctx, req)
		if err != nil {
			returnErr = apperrors.ErrInternal()
			return err
		}

		if err := uc.paymentRepo.Create(txCtx, payment); err != nil {
			returnErr = apperrors.ErrInternal()
			return err
		}

		responseBody, err := json.Marshal(payment)
		if err != nil {
			returnErr = apperrors.ErrInternal()
			return err
		}

		newRecord.Status = domain.IdempotencyStatusCompleted
		newRecord.PaymentID = payment.ID
		newRecord.ResponseBody = responseBody
		if err := uc.idempotencyRepo.Update(txCtx, newRecord); err != nil {
			returnErr = apperrors.ErrInternal()
			return err
		}

		result = &CreatePaymentResult{Payment: payment, Replayed: false}
		return nil
	})

	if txErr != nil && returnErr != nil {
		return nil, returnErr
	}
	if txErr != nil {
		return nil, apperrors.ErrInternal()
	}

	return result, nil
}

func validateIdempotencyKey(key string) error {
	if key == "" {
		return apperrors.ErrIdempotencyKeyMissing()
	}
	if len(key) > 64 {
		return apperrors.ErrIdempotencyKeyTooLong()
	}
	return nil
}

func validatePaymentRequest(req domain.PaymentRequest) error {
	var reasons []string

	if req.Amount <= 0 {
		reasons = append(reasons, "amount must be greater than 0")
	}
	if req.Currency == "" {
		reasons = append(reasons, "currency is required")
	} else if !domain.ValidCurrencies[req.Currency] {
		return apperrors.ErrInvalidCurrency(string(req.Currency))
	}
	if req.CustomerID == "" {
		reasons = append(reasons, "customer_id is required")
	}
	if req.RideID == "" {
		reasons = append(reasons, "ride_id is required")
	}
	if req.CardNumber == "" {
		reasons = append(reasons, "card_number is required")
	}

	if len(reasons) > 0 {
		return apperrors.ErrInvalidPaymentRequest(reasons[0])
	}
	return nil
}
