package use_cases

import (
	"testing"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	apperrors "github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain/errors"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/utils/fingerprint"
	"github.com/stretchr/testify/assert"
)

func validRequest() domain.PaymentRequest {
	return domain.PaymentRequest{
		Amount:      85000,
		Currency:    domain.CurrencyIDR,
		CustomerID:  "cust_001",
		RideID:      "ride_001",
		CardNumber:  "4111111111111111",
		Description: "test ride",
	}
}

func TestValidateIdempotencyKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
		errCode string
	}{
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
			errCode: "IDEMPOTENCY_KEY_MISSING",
		},
		{
			name:    "valid key",
			key:     "test-key-001",
			wantErr: false,
		},
		{
			name:    "key too long",
			key:     "aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffffffffff12345",
			wantErr: true,
			errCode: "IDEMPOTENCY_KEY_TOO_LONG",
		},
		{
			name:    "max length key",
			key:     "aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffffffffff1234",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIdempotencyKey(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
				appErr, ok := err.(*apperrors.AppError)
				assert.True(t, ok)
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePaymentRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     domain.PaymentRequest
		wantErr bool
		errCode string
	}{
		{
			name:    "valid request",
			req:     validRequest(),
			wantErr: false,
		},
		{
			name: "zero amount",
			req: domain.PaymentRequest{
				Amount:     0,
				Currency:   domain.CurrencyIDR,
				CustomerID: "cust_001",
				RideID:     "ride_001",
				CardNumber: "4111111111111111",
			},
			wantErr: true,
			errCode: "INVALID_PAYMENT_REQUEST",
		},
		{
			name: "negative amount",
			req: domain.PaymentRequest{
				Amount:     -100,
				Currency:   domain.CurrencyIDR,
				CustomerID: "cust_001",
				RideID:     "ride_001",
				CardNumber: "4111111111111111",
			},
			wantErr: true,
			errCode: "INVALID_PAYMENT_REQUEST",
		},
		{
			name: "invalid currency",
			req: domain.PaymentRequest{
				Amount:     100,
				Currency:   "USD",
				CustomerID: "cust_001",
				RideID:     "ride_001",
				CardNumber: "4111111111111111",
			},
			wantErr: true,
			errCode: "INVALID_CURRENCY",
		},
		{
			name: "missing customer_id",
			req: domain.PaymentRequest{
				Amount:     100,
				Currency:   domain.CurrencyIDR,
				CustomerID: "",
				RideID:     "ride_001",
				CardNumber: "4111111111111111",
			},
			wantErr: true,
			errCode: "INVALID_PAYMENT_REQUEST",
		},
		{
			name: "missing ride_id",
			req: domain.PaymentRequest{
				Amount:     100,
				Currency:   domain.CurrencyIDR,
				CustomerID: "cust_001",
				RideID:     "",
				CardNumber: "4111111111111111",
			},
			wantErr: true,
			errCode: "INVALID_PAYMENT_REQUEST",
		},
		{
			name: "missing card_number",
			req: domain.PaymentRequest{
				Amount:     100,
				Currency:   domain.CurrencyIDR,
				CustomerID: "cust_001",
				RideID:     "ride_001",
				CardNumber: "",
			},
			wantErr: true,
			errCode: "INVALID_PAYMENT_REQUEST",
		},
		{
			name: "missing currency",
			req: domain.PaymentRequest{
				Amount:     100,
				Currency:   "",
				CustomerID: "cust_001",
				RideID:     "ride_001",
				CardNumber: "4111111111111111",
			},
			wantErr: true,
			errCode: "INVALID_PAYMENT_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePaymentRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				appErr, ok := err.(*apperrors.AppError)
				assert.True(t, ok)
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFingerprintConsistency(t *testing.T) {
	req := validRequest()
	fp1 := fingerprint.Compute(req)
	fp2 := fingerprint.Compute(req)
	assert.Equal(t, fp1, fp2)

	req2 := validRequest()
	req2.Amount = 99999
	fp3 := fingerprint.Compute(req2)
	assert.NotEqual(t, fp1, fp3)
}
