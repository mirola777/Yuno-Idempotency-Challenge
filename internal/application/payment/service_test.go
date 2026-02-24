package payment

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	"github.com/mirola777/Yuno-Idempotency-Challenge/utils/fingerprint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type mockIdempotencyRepo struct {
	mock.Mock
}

func (m *mockIdempotencyRepo) FindByKey(ctx context.Context, key string) (*domain.IdempotencyRecord, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.IdempotencyRecord), args.Error(1)
}

func (m *mockIdempotencyRepo) Create(ctx context.Context, record *domain.IdempotencyRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *mockIdempotencyRepo) Update(ctx context.Context, record *domain.IdempotencyRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *mockIdempotencyRepo) DeleteExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockIdempotencyRepo) FindByKeyForUpdate(ctx context.Context, tx *gorm.DB, key string) (*domain.IdempotencyRecord, error) {
	args := m.Called(ctx, tx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.IdempotencyRecord), args.Error(1)
}

func (m *mockIdempotencyRepo) CreateInTx(ctx context.Context, tx *gorm.DB, record *domain.IdempotencyRecord) error {
	args := m.Called(ctx, tx, record)
	return args.Error(0)
}

func (m *mockIdempotencyRepo) UpdateInTx(ctx context.Context, tx *gorm.DB, record *domain.IdempotencyRecord) error {
	args := m.Called(ctx, tx, record)
	return args.Error(0)
}

type mockPaymentRepo struct {
	mock.Mock
}

func (m *mockPaymentRepo) Create(ctx context.Context, payment *domain.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *mockPaymentRepo) FindByID(ctx context.Context, id string) (*domain.Payment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

func (m *mockPaymentRepo) CreateInTx(ctx context.Context, tx *gorm.DB, payment *domain.Payment) error {
	args := m.Called(ctx, tx, payment)
	return args.Error(0)
}

type mockProcessor struct {
	mock.Mock
}

func (m *mockProcessor) Process(ctx context.Context, req domain.PaymentRequest) (*domain.Payment, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

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
				appErr, ok := err.(*domain.AppError)
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
				appErr, ok := err.(*domain.AppError)
				assert.True(t, ok)
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetPayment_Found(t *testing.T) {
	paymentRepo := new(mockPaymentRepo)
	expected := &domain.Payment{
		ID:     "pay-123",
		Amount: 85000,
		Status: domain.PaymentStatusSucceeded,
	}
	paymentRepo.On("FindByID", mock.Anything, "pay-123").Return(expected, nil)

	svc := &service{paymentRepo: paymentRepo}
	result, err := svc.GetPayment(context.Background(), "pay-123")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	paymentRepo.AssertExpectations(t)
}

func TestGetPayment_NotFound(t *testing.T) {
	paymentRepo := new(mockPaymentRepo)
	paymentRepo.On("FindByID", mock.Anything, "pay-unknown").Return(nil, nil)

	svc := &service{paymentRepo: paymentRepo}
	result, err := svc.GetPayment(context.Background(), "pay-unknown")

	assert.Nil(t, result)
	assert.Error(t, err)
	appErr, ok := err.(*domain.AppError)
	assert.True(t, ok)
	assert.Equal(t, "PAYMENT_NOT_FOUND", appErr.Code)
	paymentRepo.AssertExpectations(t)
}

func TestGetByIdempotencyKey_Found(t *testing.T) {
	idempotencyRepo := new(mockIdempotencyRepo)
	expected := &domain.IdempotencyRecord{
		Key:    "test-key",
		Status: domain.IdempotencyStatusCompleted,
	}
	idempotencyRepo.On("FindByKey", mock.Anything, "test-key").Return(expected, nil)

	svc := &service{idempotencyRepo: idempotencyRepo}
	result, err := svc.GetByIdempotencyKey(context.Background(), "test-key")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	idempotencyRepo.AssertExpectations(t)
}

func TestGetByIdempotencyKey_NotFound(t *testing.T) {
	idempotencyRepo := new(mockIdempotencyRepo)
	idempotencyRepo.On("FindByKey", mock.Anything, "unknown-key").Return(nil, nil)

	svc := &service{idempotencyRepo: idempotencyRepo}
	result, err := svc.GetByIdempotencyKey(context.Background(), "unknown-key")

	assert.Nil(t, result)
	assert.Error(t, err)
	appErr, ok := err.(*domain.AppError)
	assert.True(t, ok)
	assert.Equal(t, "IDEMPOTENCY_KEY_NOT_FOUND", appErr.Code)
	idempotencyRepo.AssertExpectations(t)
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

func TestFingerprintConflictDetection(t *testing.T) {
	req1 := validRequest()
	req2 := validRequest()
	req2.Amount = 200000

	fp1 := fingerprint.Compute(req1)
	fp2 := fingerprint.Compute(req2)
	assert.NotEqual(t, fp1, fp2)
}

func TestIdempotencyRecordExpiration(t *testing.T) {
	record := &domain.IdempotencyRecord{
		Key:       "test-key",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	assert.True(t, time.Now().After(record.ExpiresAt))

	activeRecord := &domain.IdempotencyRecord{
		Key:       "test-key",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	assert.True(t, time.Now().Before(activeRecord.ExpiresAt))
}

func TestPaymentResponseSerialization(t *testing.T) {
	payment := &domain.Payment{
		ID:         "pay-123",
		Amount:     85000,
		Currency:   domain.CurrencyIDR,
		CustomerID: "cust_001",
		RideID:     "ride_001",
		Status:     domain.PaymentStatusSucceeded,
		CardLast4:  "1111",
		CreatedAt:  time.Now(),
	}

	data, err := json.Marshal(payment)
	assert.NoError(t, err)

	var deserialized domain.Payment
	err = json.Unmarshal(data, &deserialized)
	assert.NoError(t, err)
	assert.Equal(t, payment.ID, deserialized.ID)
	assert.Equal(t, payment.Amount, deserialized.Amount)
	assert.Equal(t, payment.Currency, deserialized.Currency)
	assert.Equal(t, payment.Status, deserialized.Status)
}
