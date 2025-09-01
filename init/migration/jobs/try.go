package jobs

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// SeedExample inserts the example user
func SeedExample() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "sdjhbvshdbvjhdsbgjhsdbg",
		Migrate: func(db *gorm.DB) error {
			return nil
			// return db.Create(&superAdmin).Error
		},
		Rollback: func(db *gorm.DB) error {
			return nil
			// return db.Delete(&superAdmin).Error
		},
	}
}
