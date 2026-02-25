package main

import (
	"context"
	"log"
	"sl651-platform/internal/database"
	"sl651-platform/internal/model"
	"sl651-platform/internal/storage"
)

func main() {
	db, err := database.Init("data/platform.db")
	if err != nil {
		log.Fatalf("Failed to init db: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate db: %v", err)
	}

	s := storage.NewStorage(db)
	ctx := context.Background()

	// Create a test forwarding rule
	rule := &model.ForwardRule{
		ID:       "rule-v2-test",
		TenantID: "tenant-default",
		Name:     "SLT 324 SQLite Test",
		Enabled:  true,
		Filter: model.RuleFilter{
			DataTypes: []model.DataType{model.DataTypeRealtime},
		},
		Destinations: model.Destinations{
			{
				DestType: model.DestTypeDatabase,
				URL:      "data/forward_target.db", // Target SQLite DB
			},
			{
				DestType: model.DestTypeHttp,
				URL:      "http://localhost:8081/webhook", // Example HTTP center
			},
		},
	}

	if err := s.SaveForwardRule(ctx, rule); err != nil {
		log.Fatalf("Failed to save rule: %v", err)
	}

	log.Println("Test forwarding rule initialized in platform.db")
}
