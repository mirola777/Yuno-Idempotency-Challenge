package processor

import (
	"context"
	"testing"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestProcess_CardOutcomes(t *testing.T) {
	tests := []struct {
		name       string
		card       string
		wantStatus domain.PaymentStatus
		wantReason string
	}{
		{
			name:       "visa success card",
			card:       "4111111111111111",
			wantStatus: domain.PaymentStatusSucceeded,
			wantReason: "",
		},
		{
			name:       "mastercard success card",
			card:       "5500000000000004",
			wantStatus: domain.PaymentStatusSucceeded,
			wantReason: "",
		},
		{
			name:       "insufficient funds card",
			card:       "4000000000000002",
			wantStatus: domain.PaymentStatusFailed,
			wantReason: "insufficient_funds",
		},
		{
			name:       "expired card",
			card:       "4000000000000069",
			wantStatus: domain.PaymentStatusFailed,
			wantReason: "expired_card",
		},
		{
			name:       "processing error card",
			card:       "4000000000000119",
			wantStatus: domain.PaymentStatusFailed,
			wantReason: "processing_error",
		},
		{
			name:       "pending card",
			card:       "4000000000000259",
			wantStatus: domain.PaymentStatusPending,
			wantReason: "",
		},
		{
			name:       "unknown card defaults to succeeded",
			card:       "9999999999999999",
			wantStatus: domain.PaymentStatusSucceeded,
			wantReason: "",
		},
	}

	sim := NewSimulator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment, err := sim.Process(context.Background(), domain.PaymentRequest{
				Amount:     100.0,
				Currency:   domain.CurrencyIDR,
				CustomerID: "cust-1",
				RideID:     "ride-1",
				CardNumber: tt.card,
			})

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, payment.Status)
			assert.Equal(t, tt.wantReason, payment.FailReason)
		})
	}
}

func TestExtractLast4(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "16 digit card", input: "4111111111111111", want: "1111"},
		{name: "4 digit input", input: "1234", want: "1234"},
		{name: "3 digit input", input: "123", want: "123"},
		{name: "1 digit input", input: "7", want: "7"},
		{name: "empty string", input: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, extractLast4(tt.input))
		})
	}
}

func TestProcess_PaymentHasUUIDFormatID(t *testing.T) {
	sim := NewSimulator()
	payment, err := sim.Process(context.Background(), domain.PaymentRequest{
		Amount:     50.0,
		Currency:   domain.CurrencyTHB,
		CustomerID: "cust-1",
		CardNumber: "4111111111111111",
	})

	assert.NoError(t, err)
	assert.Len(t, payment.ID, 36)
	assert.Contains(t, payment.ID, "-")
}

func TestProcess_CardLast4Extracted(t *testing.T) {
	sim := NewSimulator()
	payment, err := sim.Process(context.Background(), domain.PaymentRequest{
		Amount:     100.0,
		Currency:   domain.CurrencyIDR,
		CustomerID: "cust-1",
		CardNumber: "4111111111111111",
	})

	assert.NoError(t, err)
	assert.Equal(t, "1111", payment.CardLast4)
}

func TestProcess_CurrencyPassedThrough(t *testing.T) {
	currencies := []domain.Currency{
		domain.CurrencyIDR,
		domain.CurrencyTHB,
		domain.CurrencyVND,
		domain.CurrencyPHP,
	}

	sim := NewSimulator()

	for _, cur := range currencies {
		t.Run(string(cur), func(t *testing.T) {
			payment, err := sim.Process(context.Background(), domain.PaymentRequest{
				Amount:     200.0,
				Currency:   cur,
				CustomerID: "cust-1",
				CardNumber: "4111111111111111",
			})

			assert.NoError(t, err)
			assert.Equal(t, cur, payment.Currency)
		})
	}
}

func TestProcess_AmountPassedThrough(t *testing.T) {
	amounts := []float64{0.01, 1.0, 999.99, 100000.0}

	sim := NewSimulator()

	for _, amt := range amounts {
		t.Run("", func(t *testing.T) {
			payment, err := sim.Process(context.Background(), domain.PaymentRequest{
				Amount:     amt,
				Currency:   domain.CurrencyIDR,
				CustomerID: "cust-1",
				CardNumber: "4111111111111111",
			})

			assert.NoError(t, err)
			assert.Equal(t, amt, payment.Amount)
		})
	}
}
