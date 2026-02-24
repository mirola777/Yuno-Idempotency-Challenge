package integration

import (
	"context"
	"testing"
	"time"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/application/use_cases"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	apperrors "github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain/errors"
	gormdb "github.com/mirola777/Yuno-Idempotency-Challenge/internal/infrastructure/gorm"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/infrastructure/gorm/repositories"
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
	db, err := gormdb.NewTestConnection()
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

func TestCreatePayment_ConflictDifferentCurrency(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req1 := validRequest()
	req1.Currency = domain.CurrencyTHB
	_, err := env.createPayment.Execute(ctx, "conflict-currency-key", req1)
	require.NoError(t, err)

	req2 := validRequest()
	req2.Currency = domain.CurrencyVND

	_, err = env.createPayment.Execute(ctx, "conflict-currency-key", req2)
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "IDEMPOTENCY_KEY_CONFLICT", appErr.Code)
}

func TestCreatePayment_ConflictDifferentCustomer(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req1 := validRequest()
	_, err := env.createPayment.Execute(ctx, "conflict-cust-key", req1)
	require.NoError(t, err)

	req2 := validRequest()
	req2.CustomerID = "cust-different"

	_, err = env.createPayment.Execute(ctx, "conflict-cust-key", req2)
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "IDEMPOTENCY_KEY_CONFLICT", appErr.Code)
}

func TestCreatePayment_FailedPayment_InsufficientFunds(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req := validRequest()
	req.CardNumber = "4000000000000002"

	payment, err := env.createPayment.Execute(ctx, "failed-key-funds", req)
	require.NoError(t, err)
	require.NotNil(t, payment)

	assert.Equal(t, domain.PaymentStatusFailed, payment.Status)
	assert.Equal(t, "insufficient_funds", payment.FailReason)
}

func TestCreatePayment_FailedPayment_ExpiredCard(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req := validRequest()
	req.CardNumber = "4000000000000069"

	payment, err := env.createPayment.Execute(ctx, "failed-key-expired", req)
	require.NoError(t, err)
	require.NotNil(t, payment)

	assert.Equal(t, domain.PaymentStatusFailed, payment.Status)
	assert.Equal(t, "expired_card", payment.FailReason)
}

func TestCreatePayment_FailedPayment_ProcessingError(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req := validRequest()
	req.CardNumber = "4000000000000119"

	payment, err := env.createPayment.Execute(ctx, "failed-key-processing", req)
	require.NoError(t, err)
	require.NotNil(t, payment)

	assert.Equal(t, domain.PaymentStatusFailed, payment.Status)
	assert.Equal(t, "processing_error", payment.FailReason)
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
	assert.Equal(t, 400, appErr.HTTPCode)
}

func TestCreatePayment_IdempotencyKeyTooLong(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	longKey := "aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffffffffff12345"
	_, err := env.createPayment.Execute(ctx, longKey, validRequest())
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "IDEMPOTENCY_KEY_TOO_LONG", appErr.Code)
}

func TestCreatePayment_InvalidRequest_ZeroAmount(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req := validRequest()
	req.Amount = 0

	_, err := env.createPayment.Execute(ctx, "invalid-amount-key", req)
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INVALID_PAYMENT_REQUEST", appErr.Code)
}

func TestCreatePayment_InvalidRequest_NegativeAmount(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req := validRequest()
	req.Amount = -500

	_, err := env.createPayment.Execute(ctx, "invalid-neg-key", req)
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INVALID_PAYMENT_REQUEST", appErr.Code)
}

func TestCreatePayment_InvalidCurrency(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req := validRequest()
	req.Currency = "USD"

	_, err := env.createPayment.Execute(ctx, "invalid-currency-key", req)
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INVALID_CURRENCY", appErr.Code)
}

func TestCreatePayment_MissingCustomerID(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req := validRequest()
	req.CustomerID = ""

	_, err := env.createPayment.Execute(ctx, "missing-custid-key", req)
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INVALID_PAYMENT_REQUEST", appErr.Code)
}

func TestCreatePayment_MissingRideID(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req := validRequest()
	req.RideID = ""

	_, err := env.createPayment.Execute(ctx, "missing-rideid-key", req)
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INVALID_PAYMENT_REQUEST", appErr.Code)
}

func TestCreatePayment_MissingCardNumber(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req := validRequest()
	req.CardNumber = ""

	_, err := env.createPayment.Execute(ctx, "missing-card-key", req)
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INVALID_PAYMENT_REQUEST", appErr.Code)
}

func TestCreatePayment_AllCurrencies(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	currencies := []struct {
		currency domain.Currency
		amount   float64
	}{
		{domain.CurrencyIDR, 85000},
		{domain.CurrencyTHB, 450},
		{domain.CurrencyVND, 250000},
		{domain.CurrencyPHP, 350},
	}

	for _, tc := range currencies {
		t.Run(string(tc.currency), func(t *testing.T) {
			req := validRequest()
			req.Currency = tc.currency
			req.Amount = tc.amount

			payment, err := env.createPayment.Execute(ctx, "currency-"+string(tc.currency), req)
			require.NoError(t, err)
			require.NotNil(t, payment)

			assert.Equal(t, tc.currency, payment.Currency)
			assert.Equal(t, tc.amount, payment.Amount)
			assert.Equal(t, domain.PaymentStatusSucceeded, payment.Status)
		})
	}
}

func TestCreatePayment_MultipleRetries(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()
	req := validRequest()

	var firstID string
	for i := 0; i < 5; i++ {
		payment, err := env.createPayment.Execute(ctx, "retry-key-5x", req)
		require.NoError(t, err)
		require.NotNil(t, payment)

		if i == 0 {
			firstID = payment.ID
		}
		assert.Equal(t, firstID, payment.ID)
		assert.Equal(t, domain.PaymentStatusSucceeded, payment.Status)
	}

	record, err := env.getByIdempotencyKey.Execute(ctx, "retry-key-5x")
	require.NoError(t, err)
	assert.Equal(t, firstID, record.PaymentID)
	assert.Equal(t, domain.IdempotencyStatusCompleted, record.Status)
}

func TestCreatePayment_FailedPaymentIdempotentReplay(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req := validRequest()
	req.CardNumber = "4000000000000002"

	first, err := env.createPayment.Execute(ctx, "failed-replay-key", req)
	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusFailed, first.Status)

	second, err := env.createPayment.Execute(ctx, "failed-replay-key", req)
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, domain.PaymentStatusFailed, second.Status)
	assert.Equal(t, "insufficient_funds", second.FailReason)
}

func TestCreatePayment_PendingPaymentIdempotentReplay(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req := validRequest()
	req.CardNumber = "4000000000000259"

	first, err := env.createPayment.Execute(ctx, "pending-replay-key", req)
	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusPending, first.Status)

	second, err := env.createPayment.Execute(ctx, "pending-replay-key", req)
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, domain.PaymentStatusPending, second.Status)
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
	assert.Equal(t, created.Currency, found.Currency)
	assert.Equal(t, created.CustomerID, found.CustomerID)
}

func TestGetPayment_NotFound(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	_, err := env.getPayment.Execute(ctx, "nonexistent-payment-id")
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "PAYMENT_NOT_FOUND", appErr.Code)
	assert.Equal(t, 404, appErr.HTTPCode)
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
	assert.NotEmpty(t, record.RequestFingerprint)
	assert.False(t, record.ExpiresAt.IsZero())
}

func TestGetByIdempotencyKey_NotFound(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	_, err := env.getByIdempotencyKey.Execute(ctx, "nonexistent-key")
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "IDEMPOTENCY_KEY_NOT_FOUND", appErr.Code)
	assert.Equal(t, 404, appErr.HTTPCode)
}

func TestCreatePayment_DifferentKeysCreateDifferentPayments(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()
	req := validRequest()

	p1, err := env.createPayment.Execute(ctx, "unique-key-a", req)
	require.NoError(t, err)

	p2, err := env.createPayment.Execute(ctx, "unique-key-b", req)
	require.NoError(t, err)

	assert.NotEqual(t, p1.ID, p2.ID)
}

func TestCreatePayment_DescriptionOptional(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req := validRequest()
	req.Description = ""

	payment, err := env.createPayment.Execute(ctx, "no-desc-key", req)
	require.NoError(t, err)
	require.NotNil(t, payment)

	assert.Equal(t, domain.PaymentStatusSucceeded, payment.Status)
}

func TestCreatePayment_LargeAmount(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req := validRequest()
	req.Amount = 999999999.99
	req.Currency = domain.CurrencyVND

	payment, err := env.createPayment.Execute(ctx, "large-amount-key", req)
	require.NoError(t, err)
	require.NotNil(t, payment)

	assert.Equal(t, 999999999.99, payment.Amount)
}

func TestCreatePayment_CardLast4Masked(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req := validRequest()
	req.CardNumber = "4111111111111111"

	payment, err := env.createPayment.Execute(ctx, "card-mask-key", req)
	require.NoError(t, err)

	assert.Equal(t, "1111", payment.CardLast4)
}

func TestEndToEnd_FullPaymentLifecycle(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	req := domain.PaymentRequest{
		Amount:      85000,
		Currency:    domain.CurrencyIDR,
		CustomerID:  "cust_jakarta_001",
		RideID:      "ride_jkt_001",
		CardNumber:  "4111111111111111",
		Description: "Ride from Sudirman to Kemang",
	}

	payment, err := env.createPayment.Execute(ctx, "lifecycle-key", req)
	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusSucceeded, payment.Status)

	replay, err := env.createPayment.Execute(ctx, "lifecycle-key", req)
	require.NoError(t, err)
	assert.Equal(t, payment.ID, replay.ID)

	found, err := env.getPayment.Execute(ctx, payment.ID)
	require.NoError(t, err)
	assert.Equal(t, payment.ID, found.ID)

	record, err := env.getByIdempotencyKey.Execute(ctx, "lifecycle-key")
	require.NoError(t, err)
	assert.Equal(t, domain.IdempotencyStatusCompleted, record.Status)
	assert.Equal(t, payment.ID, record.PaymentID)
}
