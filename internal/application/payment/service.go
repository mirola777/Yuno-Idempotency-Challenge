package payment

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	"github.com/mirola777/Yuno-Idempotency-Challenge/utils/fingerprint"
	"gorm.io/gorm"
)

type service struct {
	db              *gorm.DB
	idempotencyRepo domain.IdempotencyRepository
	paymentRepo     domain.PaymentRepository
	processor       domain.PaymentProcessor
	keyTTL          time.Duration
}

func NewService(
	db *gorm.DB,
	idempotencyRepo domain.IdempotencyRepository,
	paymentRepo domain.PaymentRepository,
	processor domain.PaymentProcessor,
	keyTTL time.Duration,
) domain.PaymentService {
	return &service{
		db:              db,
		idempotencyRepo: idempotencyRepo,
		paymentRepo:     paymentRepo,
		processor:       processor,
		keyTTL:          keyTTL,
	}
}

func (s *service) CreatePayment(ctx context.Context, idempotencyKey string, req domain.PaymentRequest) (*domain.Payment, error) {
	if err := validateIdempotencyKey(idempotencyKey); err != nil {
		return nil, err
	}

	if err := validatePaymentRequest(req); err != nil {
		return nil, err
	}

	fp := fingerprint.Compute(req)

	var result *domain.Payment
	var returnErr error

	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		record, err := s.idempotencyRepo.FindByKeyForUpdate(ctx, tx, idempotencyKey)
		if err != nil {
			returnErr = domain.ErrInternal("failed to check idempotency key")
			return err
		}

		if record != nil {
			if record.Status == domain.IdempotencyStatusProcessing {
				returnErr = domain.ErrPaymentProcessing()
				return returnErr
			}

			if record.RequestFingerprint != fp {
				returnErr = domain.ErrIdempotencyKeyConflict(idempotencyKey)
				return returnErr
			}

			var cached domain.Payment
			if err := json.Unmarshal(record.ResponseBody, &cached); err != nil {
				returnErr = domain.ErrInternal("failed to deserialize cached response")
				return err
			}
			result = &cached
			return nil
		}

		newRecord := &domain.IdempotencyRecord{
			Key:                idempotencyKey,
			RequestFingerprint: fp,
			Status:             domain.IdempotencyStatusProcessing,
			CreatedAt:          time.Now(),
			ExpiresAt:          time.Now().Add(s.keyTTL),
		}
		if err := s.idempotencyRepo.CreateInTx(ctx, tx, newRecord); err != nil {
			returnErr = domain.ErrInternal("failed to create idempotency record")
			return err
		}

		payment, err := s.processor.Process(ctx, req)
		if err != nil {
			returnErr = domain.ErrInternal("payment processing failed")
			return err
		}

		if err := s.paymentRepo.CreateInTx(ctx, tx, payment); err != nil {
			returnErr = domain.ErrInternal("failed to save payment")
			return err
		}

		responseBody, err := json.Marshal(payment)
		if err != nil {
			returnErr = domain.ErrInternal("failed to serialize payment response")
			return err
		}

		newRecord.Status = domain.IdempotencyStatusCompleted
		newRecord.PaymentID = payment.ID
		newRecord.ResponseBody = responseBody
		if err := s.idempotencyRepo.UpdateInTx(ctx, tx, newRecord); err != nil {
			returnErr = domain.ErrInternal("failed to update idempotency record")
			return err
		}

		result = payment
		return nil
	})

	if txErr != nil && returnErr != nil {
		return nil, returnErr
	}
	if txErr != nil {
		return nil, domain.ErrInternal("transaction failed")
	}

	return result, nil
}

func (s *service) GetPayment(ctx context.Context, paymentID string) (*domain.Payment, error) {
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, domain.ErrInternal("failed to retrieve payment")
	}
	if payment == nil {
		return nil, domain.ErrPaymentNotFound(paymentID)
	}
	return payment, nil
}

func (s *service) GetByIdempotencyKey(ctx context.Context, key string) (*domain.IdempotencyRecord, error) {
	record, err := s.idempotencyRepo.FindByKey(ctx, key)
	if err != nil {
		return nil, domain.ErrInternal("failed to retrieve idempotency record")
	}
	if record == nil {
		return nil, domain.ErrIdempotencyKeyNotFound(key)
	}
	return record, nil
}

func validateIdempotencyKey(key string) error {
	if key == "" {
		return domain.ErrIdempotencyKeyMissing()
	}
	if len(key) > 64 {
		return domain.ErrIdempotencyKeyTooLong()
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
		return domain.ErrInvalidCurrency(string(req.Currency))
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
		return domain.ErrInvalidPaymentRequest(reasons)
	}
	return nil
}
