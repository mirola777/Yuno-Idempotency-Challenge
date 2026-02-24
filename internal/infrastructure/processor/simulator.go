package processor

import (
	"context"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
)

type Simulator struct{}

func NewSimulator() domain.PaymentProcessor {
	return &Simulator{}
}

func (s *Simulator) Process(_ context.Context, req domain.PaymentRequest) (*domain.Payment, error) {
	delay := time.Duration(50+rand.Intn(150)) * time.Millisecond
	time.Sleep(delay)

	status, failReason := resolveOutcome(req.CardNumber)
	cardLast4 := extractLast4(req.CardNumber)

	return &domain.Payment{
		ID:          uuid.New().String(),
		Amount:      req.Amount,
		Currency:    req.Currency,
		CustomerID:  req.CustomerID,
		RideID:      req.RideID,
		Status:      status,
		CardLast4:   cardLast4,
		Description: req.Description,
		FailReason:  failReason,
		CreatedAt:   time.Now(),
	}, nil
}

func resolveOutcome(cardNumber string) (domain.PaymentStatus, string) {
	switch cardNumber {
	case "4000000000000002":
		return domain.PaymentStatusFailed, "insufficient_funds"
	case "4000000000000069":
		return domain.PaymentStatusFailed, "expired_card"
	case "4000000000000119":
		return domain.PaymentStatusFailed, "processing_error"
	case "4000000000000259":
		return domain.PaymentStatusPending, ""
	default:
		return domain.PaymentStatusSucceeded, ""
	}
}

func extractLast4(cardNumber string) string {
	if len(cardNumber) < 4 {
		return cardNumber
	}
	return cardNumber[len(cardNumber)-4:]
}
