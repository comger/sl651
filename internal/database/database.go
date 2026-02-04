package database

import (
	"os"
	"path/filepath"
	"sl651-platform/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Init(path string) (*gorm.DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Device{},
		&model.DeviceData{},
		&model.Tenant{},
		&model.ForwardRule{},
		&model.ForwardLog{},
		&model.StatusHistory{},
		&model.FaultLog{},
		&model.SystemLog{},
		&model.QualityMetric{},
		&model.DeviceCommand{},
	)
}
