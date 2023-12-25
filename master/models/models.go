package models

import (
	"log"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&User{},
		&Account{},
		&RouteTask{},
		&Fleet{},
		&taskLog{},
	)
	if err != nil {
		log.Fatalf("Error during migration: %v", err)
	}
}
