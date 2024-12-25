package models

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&User{},
		&Account{},
		&Fleet{},
		&Task{},
		&Target{},
		&TaskLog{},
	)
	if err != nil {
		log.Fatal("Error during migration: %v",
			zap.Error(err))
	}
}
