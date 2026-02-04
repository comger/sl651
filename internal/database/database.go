package database

import (
	"sl651-platform/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Init(path string) (*gorm.DB, error) {
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
	)
}
