package migration

import (
	"fmt"

	"github.com/Bendomey/fincore-engine/init/migration/jobs"
	log "github.com/sirupsen/logrus"

	"github.com/Bendomey/fincore-engine/internal/models"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func updateMigration(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Client{},
	)
	return err
}

// ServiceAutoMigration migrates all the tables and modifications to the connected source
func ServiceAutoMigration(db *gorm.DB) error {
	// Keep a list of migrations here
	m := gormigrate.New(db, gormigrate.DefaultOptions, nil)

	m.InitSchema(func(db *gorm.DB) error {
		log.Info("[Migration.InitSchema] Initializing database schema")
		db.Exec("create extension \"uuid-ossp\";")
		if err := updateMigration(db); err != nil {
			return fmt.Errorf("[Migration.InitSchema]: %v", err)
		}
		// Add more jobs, etc here
		return nil
	})
	m.Migrate()

	if err := updateMigration(db); err != nil {
		return err
	}

	m = gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		jobs.SeedExample(),
	})
	m.Migrate()

	return nil
}
