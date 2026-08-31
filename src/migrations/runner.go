package migrations

import (
	"fmt"
	"time"

	"bazi/internal/model"

	"gorm.io/gorm"
)

const CurrentVersion = 1

type schemaMigration struct {
	Version   int       `gorm:"primaryKey;autoIncrement:false"`
	Name      string    `gorm:"type:varchar(128);not null"`
	AppliedAt time.Time `gorm:"not null"`
}

func (schemaMigration) TableName() string {
	return "schema_migrations"
}

type migrationStep struct {
	version int
	name    string
	apply   func(*gorm.DB) error
}

var migrationSteps = []migrationStep{
	{
		version: 1,
		name:    "baseline_models",
		apply: func(db *gorm.DB) error {
			return db.AutoMigrate(
				&model.User{},
				&model.BirthChart{},
				&model.FortuneRecord{},
				&model.FortuneFeedback{},
				&model.BuyiRecord{},
				&model.ElementAsset{},
				&model.AuspiciousRule{},
				&model.ActivityCatalog{},
			)
		},
	},
}

// Apply runs pending schema migrations in order. Each completed version is
// recorded so application startup is idempotent on both SQLite and MySQL.
func Apply(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database is required")
	}
	if err := db.AutoMigrate(&schemaMigration{}); err != nil {
		return fmt.Errorf("create schema migration table: %w", err)
	}

	var latest schemaMigration
	result := db.Order("version DESC").Limit(1).Find(&latest)
	if result.Error != nil {
		return fmt.Errorf("read schema version: %w", result.Error)
	}
	if result.RowsAffected > 0 && latest.Version > CurrentVersion {
		return fmt.Errorf("database schema version %d is newer than supported version %d", latest.Version, CurrentVersion)
	}

	for _, step := range migrationSteps {
		var applied int64
		if err := db.Model(&schemaMigration{}).Where("version = ?", step.version).Count(&applied).Error; err != nil {
			return fmt.Errorf("check schema migration %d: %w", step.version, err)
		}
		if applied > 0 {
			continue
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := step.apply(tx); err != nil {
				return err
			}
			return tx.Create(&schemaMigration{
				Version:   step.version,
				Name:      step.name,
				AppliedAt: time.Now().UTC(),
			}).Error
		}); err != nil {
			return fmt.Errorf("apply schema migration %d (%s): %w", step.version, step.name, err)
		}
	}

	return nil
}
