package gormdb

import (
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/infrastructure/gorm/migrations"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	return migrations.Run(db)
}
