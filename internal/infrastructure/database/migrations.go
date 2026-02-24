package database

import (
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/infrastructure/database/migrations"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	return migrations.Run(db)
}
