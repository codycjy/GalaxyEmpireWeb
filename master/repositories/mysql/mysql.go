//go:build !test

package mysql

import (
	"GalaxyEmpireWeb/config"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	once     sync.Once
	globalDB *gorm.DB
	err      error
)

func ConnectDatabase() {
	dsn := config.GetDSN("config/yaml/mysql.yaml")
	globalDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	db, _ := globalDB.DB()
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(50)
	db.SetConnMaxLifetime(time.Hour)
	if err != nil {
		panic("failed to connect database")
	}
}

func GetDB() *gorm.DB {
	once.Do(func() {
		ConnectDatabase()
	})
	return globalDB
}
