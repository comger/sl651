package adapters

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"sl651-platform/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLAdapter struct {
	dbs map[string]*gorm.DB
}

func NewSQLAdapter() *SQLAdapter {
	return &SQLAdapter{
		dbs: make(map[string]*gorm.DB),
	}
}

func (a *SQLAdapter) Send(ctx context.Context, data *model.DeviceData, dest model.Destination) error {
	db, err := a.getDB(dest)
	if err != nil {
		return err
	}

	// SLT 324 Mapping Logic
	// We'll focus on common tables: ST_PPTN_R (Rain), ST_RIVER_R (River/Water Level)

	stcd := data.DeviceID
	tm := data.Timestamp

	for _, p := range data.Values {
		switch p.Tag {
		case "DRP": // Precipitation
			prec := a.valToFloat(p.Value)
			if prec != nil {
				record := map[string]interface{}{
					"STCD": stcd,
					"TM":   tm,
					"DRP":  *prec,
				}
				if err := db.Table("ST_PPTN_R").Create(record).Error; err != nil {
					log.Printf("Failed to insert into ST_PPTN_R: %v", err)
				}
			}
		case "Z": // Water Level
			level := a.valToFloat(p.Value)
			if level != nil {
				record := map[string]interface{}{
					"STCD": stcd,
					"TM":   tm,
					"Z":    *level,
				}
				// Also check for Q (Flow) in same data if possible?
				// For simplicity, we just insert into ST_RIVER_R
				if err := db.Table("ST_RIVER_R").Create(record).Error; err != nil {
					log.Printf("Failed to insert into ST_RIVER_R: %v", err)
				}
			}
		}
	}

	return nil
}

func (a *SQLAdapter) getDB(dest model.Destination) (*gorm.DB, error) {
	dsn := dest.URL
	driver := string(dest.DestType)

	// Support direct URL format detection and driver assignment
	if strings.HasPrefix(strings.ToLower(dsn), "mysql://") {
		driver = "mysql"
		dsn = a.convertToMySQLDSN(dsn)
	} else if strings.HasPrefix(strings.ToLower(dsn), "postgres://") || strings.HasPrefix(strings.ToLower(dsn), "postgresql://") {
		driver = "postgres"
	}

	cacheKey := dsn + ":" + driver
	if db, ok := a.dbs[cacheKey]; ok {
		return db, nil
	}

	var dialector gorm.Dialector
	switch strings.ToLower(driver) {
	case "postgres", "postgresql":
		dialector = postgres.Open(dsn)
	case "mysql":
		dialector = mysql.Open(dsn)
	case "sqlite", "sqlite3":
		dialector = sqlite.Open(dsn)
	default:
		// Database or fallback guess
		if strings.Contains(dsn, "user=") || strings.Contains(dsn, "host=") {
			dialector = postgres.Open(dsn)
		} else if strings.Contains(dsn, "@tcp(") {
			dialector = mysql.Open(dsn)
		} else {
			dialector = sqlite.Open(dsn)
		}
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s database (%s): %w", driver, dsn, err)
	}

	// Auto Migrate for SLT 324 Tables if they don't exist (basic version)
	a.ensureTables(db)

	a.dbs[cacheKey] = db
	return db, nil
}

func (a *SQLAdapter) convertToMySQLDSN(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL // Return as is if unparseable
	}

	user := u.User.Username()
	pass, _ := u.User.Password()
	host := u.Host
	if !strings.Contains(host, ":") {
		host = host + ":3306" // Default port
	}

	dbName := strings.TrimPrefix(u.Path, "/")
	query := u.RawQuery

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, pass, host, dbName)
	if query != "" {
		dsn = dsn + "?" + query
	}
	return dsn
}

func (a *SQLAdapter) ensureTables(db *gorm.DB) {
	// Crude table check/creation for SLT 324 compliance
	// ST_PPTN_R: Precipitation
	// ST_RIVER_R: River Stage

	// Check if ST_PPTN_R exists
	if !db.Migrator().HasTable("ST_PPTN_R") {
		db.Exec(`CREATE TABLE ST_PPTN_R (
			STCD CHAR(8) NOT NULL,
			TM DATETIME NOT NULL,
			DRP DECIMAL(5,1),
			PRIMARY KEY (STCD, TM)
		)`)
	}
	if !db.Migrator().HasTable("ST_RIVER_R") {
		db.Exec(`CREATE TABLE ST_RIVER_R (
			STCD CHAR(8) NOT NULL,
			TM DATETIME NOT NULL,
			Z DECIMAL(7,3),
			Q DECIMAL(9,3),
			PRIMARY KEY (STCD, TM)
		)`)
	}
}

func (a *SQLAdapter) valToFloat(v model.DataValue) *float64 {
	if v.Float != nil {
		return v.Float
	}
	if v.Int != nil {
		f := float64(*v.Int)
		return &f
	}
	return nil
}

func (a *SQLAdapter) Close() error {
	for _, db := range a.dbs {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
	return nil
}
