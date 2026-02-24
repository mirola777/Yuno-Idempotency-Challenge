package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/infrastructure/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupPaymentTest(t *testing.T) (*PaymentRepo, *gorm.DB) {
	db, err := database.NewTestConnection()
	require.NoError(t, err)
	repo := &PaymentRepo{db: db}
	return repo, db
}

func TestPaymentCreate_And_FindByID(t *testing.T) {
	repo, _ := setupPaymentTest(t)
	ctx := context.Background()

	payment := &domain.Payment{
		ID:          "pay-001",
		Amount:      100.50,
		Currency:    domain.CurrencyIDR,
		CustomerID:  "cust-001",
		RideID:      "ride-001",
		Status:      domain.PaymentStatusSucceeded,
		CardLast4:   "4242",
		Description: "Test payment",
		CreatedAt:   time.Now(),
	}

	err := repo.Create(ctx, payment)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, "pay-001")
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, payment.ID, found.ID)
	assert.Equal(t, payment.Amount, found.Amount)
	assert.Equal(t, payment.Currency, found.Currency)
	assert.Equal(t, payment.CustomerID, found.CustomerID)
	assert.Equal(t, payment.RideID, found.RideID)
	assert.Equal(t, payment.Status, found.Status)
	assert.Equal(t, payment.CardLast4, found.CardLast4)
	assert.Equal(t, payment.Description, found.Description)
}

func TestPaymentFindByID_NotFound(t *testing.T) {
	repo, _ := setupPaymentTest(t)
	ctx := context.Background()

	found, err := repo.FindByID(ctx, "nonexistent-id")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestPaymentCreateInTx(t *testing.T) {
	repo, db := setupPaymentTest(t)
	ctx := context.Background()

	tx := db.Begin()
	require.NoError(t, tx.Error)

	payment := &domain.Payment{
		ID:         "pay-tx-001",
		Amount:     250.00,
		Currency:   domain.CurrencyTHB,
		CustomerID: "cust-tx-001",
		RideID:     "ride-tx-001",
		Status:     domain.PaymentStatusPending,
		CardLast4:  "0259",
		CreatedAt:  time.Now(),
	}

	err := repo.CreateInTx(ctx, tx, payment)
	require.NoError(t, err)

	err = tx.Commit().Error
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, "pay-tx-001")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "pay-tx-001", found.ID)
	assert.Equal(t, domain.PaymentStatusPending, found.Status)
}

func TestPaymentCreate_DuplicateID(t *testing.T) {
	repo, _ := setupPaymentTest(t)
	ctx := context.Background()

	payment := &domain.Payment{
		ID:         "pay-dup-001",
		Amount:     50.00,
		Currency:   domain.CurrencyVND,
		CustomerID: "cust-dup",
		RideID:     "ride-dup",
		Status:     domain.PaymentStatusSucceeded,
		CardLast4:  "1234",
		CreatedAt:  time.Now(),
	}

	err := repo.Create(ctx, payment)
	require.NoError(t, err)

	duplicate := &domain.Payment{
		ID:         "pay-dup-001",
		Amount:     75.00,
		Currency:   domain.CurrencyPHP,
		CustomerID: "cust-dup-2",
		RideID:     "ride-dup-2",
		Status:     domain.PaymentStatusFailed,
		CardLast4:  "5678",
		CreatedAt:  time.Now(),
	}

	err = repo.Create(ctx, duplicate)
	assert.Error(t, err)
}
