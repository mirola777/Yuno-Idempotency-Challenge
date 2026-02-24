package database

import (
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.Payment{},
		&domain.IdempotencyRecord{},
	)
}
