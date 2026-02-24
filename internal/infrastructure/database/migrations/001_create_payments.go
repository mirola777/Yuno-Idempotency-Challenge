package migrations

import (
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	"gorm.io/gorm"
)

func init() {
	Register(Migration{
		ID: "001_create_payments",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&domain.Payment{})
		},
	})
}
