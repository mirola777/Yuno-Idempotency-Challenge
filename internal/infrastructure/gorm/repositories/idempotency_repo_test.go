package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	gormdb "github.com/mirola777/Yuno-Idempotency-Challenge/internal/infrastructure/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupIdempotencyTest(t *testing.T) (*IdempotencyRepo, *gorm.DB) {
	db, err := gormdb.NewTestConnection()
	require.NoError(t, err)
	repo := &IdempotencyRepo{db: db}
	return repo, db
}

func TestCreate_And_FindByKey(t *testing.T) {
	repo, _ := setupIdempotencyTest(t)
	ctx := context.Background()

	record := &domain.IdempotencyRecord{
		Key:                "test-key-001",
		RequestFingerprint: "fingerprint-abc",
		PaymentID:          "pay-123",
		ResponseBody:       []byte(`{"id":"pay-123"}`),
		Status:             domain.IdempotencyStatusCompleted,
		CreatedAt:          time.Now(),
		ExpiresAt:          time.Now().Add(24 * time.Hour),
	}

	err := repo.Create(ctx, record)
	require.NoError(t, err)

	found, err := repo.FindByKey(ctx, "test-key-001")
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, record.Key, found.Key)
	assert.Equal(t, record.RequestFingerprint, found.RequestFingerprint)
	assert.Equal(t, record.PaymentID, found.PaymentID)
	assert.Equal(t, record.Status, found.Status)
	assert.Equal(t, record.ResponseBody, found.ResponseBody)
}

func TestFindByKey_NotFound(t *testing.T) {
	repo, _ := setupIdempotencyTest(t)
	ctx := context.Background()

	found, err := repo.FindByKey(ctx, "nonexistent-key")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestFindByKey_Expired(t *testing.T) {
	repo, _ := setupIdempotencyTest(t)
	ctx := context.Background()

	record := &domain.IdempotencyRecord{
		Key:                "expired-key",
		RequestFingerprint: "fingerprint-expired",
		PaymentID:          "pay-expired",
		ResponseBody:       []byte(`{"id":"pay-expired"}`),
		Status:             domain.IdempotencyStatusCompleted,
		CreatedAt:          time.Now().Add(-48 * time.Hour),
		ExpiresAt:          time.Now().Add(-1 * time.Hour),
	}

	err := repo.Create(ctx, record)
	require.NoError(t, err)

	found, err := repo.FindByKey(ctx, "expired-key")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestUpdate(t *testing.T) {
	repo, _ := setupIdempotencyTest(t)
	ctx := context.Background()

	record := &domain.IdempotencyRecord{
		Key:                "update-key",
		RequestFingerprint: "fingerprint-update",
		Status:             domain.IdempotencyStatusProcessing,
		CreatedAt:          time.Now(),
		ExpiresAt:          time.Now().Add(24 * time.Hour),
	}

	err := repo.Create(ctx, record)
	require.NoError(t, err)

	record.Status = domain.IdempotencyStatusCompleted
	record.PaymentID = "pay-updated"
	record.ResponseBody = []byte(`{"id":"pay-updated"}`)

	err = repo.Update(ctx, record)
	require.NoError(t, err)

	found, err := repo.FindByKey(ctx, "update-key")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, domain.IdempotencyStatusCompleted, found.Status)
	assert.Equal(t, "pay-updated", found.PaymentID)
}

func TestDeleteExpired(t *testing.T) {
	repo, _ := setupIdempotencyTest(t)
	ctx := context.Background()

	expired := &domain.IdempotencyRecord{
		Key:                "expired-delete-key",
		RequestFingerprint: "fp-expired",
		Status:             domain.IdempotencyStatusCompleted,
		CreatedAt:          time.Now().Add(-48 * time.Hour),
		ExpiresAt:          time.Now().Add(-1 * time.Hour),
	}
	err := repo.Create(ctx, expired)
	require.NoError(t, err)

	active := &domain.IdempotencyRecord{
		Key:                "active-key",
		RequestFingerprint: "fp-active",
		Status:             domain.IdempotencyStatusCompleted,
		CreatedAt:          time.Now(),
		ExpiresAt:          time.Now().Add(24 * time.Hour),
	}
	err = repo.Create(ctx, active)
	require.NoError(t, err)

	count, err := repo.DeleteExpired(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	found, err := repo.FindByKey(ctx, "active-key")
	require.NoError(t, err)
	assert.NotNil(t, found)
}

func TestFindByKeyForUpdate(t *testing.T) {
	repo, db := setupIdempotencyTest(t)
	ctx := context.Background()

	record := &domain.IdempotencyRecord{
		Key:                "for-update-key",
		RequestFingerprint: "fp-for-update",
		PaymentID:          "pay-for-update",
		ResponseBody:       []byte(`{"id":"pay-for-update"}`),
		Status:             domain.IdempotencyStatusCompleted,
		CreatedAt:          time.Now(),
		ExpiresAt:          time.Now().Add(24 * time.Hour),
	}

	err := repo.Create(ctx, record)
	require.NoError(t, err)

	tx := db.Begin()
	require.NoError(t, tx.Error)

	txCtx := gormdb.WithTx(ctx, tx)
	found, err := repo.FindByKeyForUpdate(txCtx, "for-update-key")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "for-update-key", found.Key)
	assert.Equal(t, domain.IdempotencyStatusCompleted, found.Status)

	tx.Commit()
}

func TestCreateInTransaction(t *testing.T) {
	repo, db := setupIdempotencyTest(t)
	ctx := context.Background()

	tx := db.Begin()
	require.NoError(t, tx.Error)

	record := &domain.IdempotencyRecord{
		Key:                "tx-create-key",
		RequestFingerprint: "fp-tx-create",
		Status:             domain.IdempotencyStatusProcessing,
		CreatedAt:          time.Now(),
		ExpiresAt:          time.Now().Add(24 * time.Hour),
	}

	txCtx := gormdb.WithTx(ctx, tx)
	err := repo.Create(txCtx, record)
	require.NoError(t, err)

	err = tx.Commit().Error
	require.NoError(t, err)

	found, err := repo.FindByKey(ctx, "tx-create-key")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "tx-create-key", found.Key)
	assert.Equal(t, domain.IdempotencyStatusProcessing, found.Status)
}

func TestUpdateInTransaction(t *testing.T) {
	repo, db := setupIdempotencyTest(t)
	ctx := context.Background()

	record := &domain.IdempotencyRecord{
		Key:                "tx-update-key",
		RequestFingerprint: "fp-tx-update",
		Status:             domain.IdempotencyStatusProcessing,
		CreatedAt:          time.Now(),
		ExpiresAt:          time.Now().Add(24 * time.Hour),
	}

	err := repo.Create(ctx, record)
	require.NoError(t, err)

	tx := db.Begin()
	require.NoError(t, tx.Error)

	record.Status = domain.IdempotencyStatusCompleted
	record.PaymentID = "pay-tx-updated"
	record.ResponseBody = []byte(`{"id":"pay-tx-updated"}`)

	txCtx := gormdb.WithTx(ctx, tx)
	err = repo.Update(txCtx, record)
	require.NoError(t, err)

	err = tx.Commit().Error
	require.NoError(t, err)

	found, err := repo.FindByKey(ctx, "tx-update-key")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, domain.IdempotencyStatusCompleted, found.Status)
	assert.Equal(t, "pay-tx-updated", found.PaymentID)
}
