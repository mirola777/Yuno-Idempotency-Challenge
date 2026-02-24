package migrations

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type Migration struct {
	ID      string
	Migrate func(tx *gorm.DB) error
}

type MigrationRecord struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"`
	MigrationID string `gorm:"uniqueIndex;not null"`
	CreatedAt   time.Time
}

func (MigrationRecord) TableName() string {
	return "schema_migrations"
}

var registry []Migration

func Register(m Migration) {
	registry = append(registry, m)
}

func Run(db *gorm.DB) error {
	db.AutoMigrate(&MigrationRecord{})

	for _, m := range registry {
		var count int64
		db.Model(&MigrationRecord{}).Where("migration_id = ?", m.ID).Count(&count)
		if count > 0 {
			continue
		}

		log.Printf("running migration: %s", m.ID)
		if err := m.Migrate(db); err != nil {
			return fmt.Errorf("migration %s failed: %w", m.ID, err)
		}

		db.Create(&MigrationRecord{MigrationID: m.ID})
		log.Printf("completed migration: %s", m.ID)
	}
	return nil
}
