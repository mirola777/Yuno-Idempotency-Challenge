package migrations

import (
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	"gorm.io/gorm"
)

func init() {
	Register(Migration{
		ID: "002_create_idempotency_records",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&domain.IdempotencyRecord{})
		},
	})
}
