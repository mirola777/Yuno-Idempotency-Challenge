package integration

import (
	"context"
	"testing"
	"time"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/application/use_cases"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	apperrors "github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain/errors"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/infrastructure/database"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/infrastructure/database/repositories"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/infrastructure/processor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testEnv struct {
	createPayment       *use_cases.CreatePaymentUseCase
	getPayment          *use_cases.GetPaymentUseCase
	getByIdempotencyKey *use_cases.GetByIdempotencyKeyUseCase
}

func setupIntegration(t *testing.T) *testEnv {
	db, err := database.NewTestConnection()
	require.NoError(t, err)

	idempotencyRepo := repositories.NewIdempotencyRepo(db)
	paymentRepo := repositories.NewPaymentRepo(db)
	paymentProcessor := processor.NewSimulator()
	keyTTL := 24 * time.Hour

	return &testEnv{
		createPayment:       use_cases.NewCreatePaymentUseCase(db, idempotencyRepo, paymentRepo, paymentProcessor, keyTTL),
		getPayment:          use_cases.NewGetPaymentUseCase(paymentRepo),
		getByIdempotencyKey: use_cases.NewGetByIdempotencyKeyUseCase(idempotencyRepo),
	}
}

func validRequest() domain.PaymentRequest {
	return domain.PaymentRequest{
		Amount:      100.00,
		Currency:    domain.CurrencyIDR,
		CustomerID:  "cust-001",
		RideID:      "ride-001",
		CardNumber:  "4242424242424242",
		Description: "Test ride payment",
	}
}

func TestCreatePayment_NewKey_Succeeds(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	payment, err := env.createPayment.Execute(ctx, "new-key-001", validRequest())
	require.NoError(t, err)
	require.NotNil(t, payment)

	assert.NotEmpty(t, payment.ID)
	assert.Equal(t, domain.PaymentStatusSucceeded, payment.Status)
	assert.Equal(t, 100.00, payment.Amount)
	assert.Equal(t, domain.CurrencyIDR, payment.Currency)
	assert.Equal(t, "cust-001", payment.CustomerID)
	assert.Equal(t, "ride-001", payment.RideID)
	assert.Equal(t, "4242", payment.CardLast4)
}

func TestCreatePayment_IdempotentReplay(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()
	req := validRequest()

	first, err := env.createPayment.Execute(ctx, "replay-key", req)
	require.NoError(t, err)
	require.NotNil(t, first)

	second, err := env.createPayment.Execute(ctx, "replay-key", req)
	require.NoError(t, err)
	require.NotNil(t, second)

	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, first.Amount, second.Amount)
	assert.Equal(t, first.Status, second.Status)
}

func TestCreatePayment_ConflictDetection(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req1 := validRequest()
	_, err := env.createPayment.Execute(ctx, "conflict-key", req1)
	require.NoError(t, err)

	req2 := validRequest()
	req2.Amount = 999.99

	_, err = env.createPayment.Execute(ctx, "conflict-key", req2)
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "IDEMPOTENCY_KEY_CONFLICT", appErr.Code)
	assert.Equal(t, 409, appErr.HTTPCode)
}

func TestCreatePayment_FailedPayment(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req := validRequest()
	req.CardNumber = "4000000000000002"

	payment, err := env.createPayment.Execute(ctx, "failed-key", req)
	require.NoError(t, err)
	require.NotNil(t, payment)

	assert.Equal(t, domain.PaymentStatusFailed, payment.Status)
	assert.Equal(t, "insufficient_funds", payment.FailReason)
}

func TestCreatePayment_PendingPayment(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req := validRequest()
	req.CardNumber = "4000000000000259"

	payment, err := env.createPayment.Execute(ctx, "pending-key", req)
	require.NoError(t, err)
	require.NotNil(t, payment)

	assert.Equal(t, domain.PaymentStatusPending, payment.Status)
	assert.Empty(t, payment.FailReason)
}

func TestCreatePayment_MissingIdempotencyKey(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	_, err := env.createPayment.Execute(ctx, "", validRequest())
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "IDEMPOTENCY_KEY_MISSING", appErr.Code)
}

func TestCreatePayment_InvalidRequest(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req := validRequest()
	req.Amount = 0

	_, err := env.createPayment.Execute(ctx, "invalid-key", req)
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INVALID_PAYMENT_REQUEST", appErr.Code)
}

func TestGetPayment_ExistingPayment(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	created, err := env.createPayment.Execute(ctx, "get-pay-key", validRequest())
	require.NoError(t, err)
	require.NotNil(t, created)

	found, err := env.getPayment.Execute(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, created.Amount, found.Amount)
	assert.Equal(t, created.Status, found.Status)
}

func TestGetPayment_NotFound(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	_, err := env.getPayment.Execute(ctx, "nonexistent-payment-id")
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "PAYMENT_NOT_FOUND", appErr.Code)
}

func TestGetByIdempotencyKey_ExistingKey(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	payment, err := env.createPayment.Execute(ctx, "lookup-key", validRequest())
	require.NoError(t, err)
	require.NotNil(t, payment)

	record, err := env.getByIdempotencyKey.Execute(ctx, "lookup-key")
	require.NoError(t, err)
	require.NotNil(t, record)

	assert.Equal(t, "lookup-key", record.Key)
	assert.Equal(t, payment.ID, record.PaymentID)
	assert.Equal(t, domain.IdempotencyStatusCompleted, record.Status)
}

func TestGetByIdempotencyKey_NotFound(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	_, err := env.getByIdempotencyKey.Execute(ctx, "nonexistent-key")
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "IDEMPOTENCY_KEY_NOT_FOUND", appErr.Code)
}

func TestCreatePayment_MultipleRetries(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()
	req := validRequest()

	var firstID string
	for i := 0; i < 5; i++ {
		payment, err := env.createPayment.Execute(ctx, "retry-key", req)
		require.NoError(t, err)
		require.NotNil(t, payment)

		if i == 0 {
			firstID = payment.ID
		}
		assert.Equal(t, firstID, payment.ID)
	}

	record, err := env.getByIdempotencyKey.Execute(ctx, "retry-key")
	require.NoError(t, err)
	assert.Equal(t, firstID, record.PaymentID)
}
