package gormdb

import (
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewTestConnection() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&domain.Payment{}, &domain.IdempotencyRecord{})
	return db, nil
}
