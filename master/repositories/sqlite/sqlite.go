//go:build !test

package sqlite

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return db
}
func GetTestDB() *gorm.DB {
	if testDB == nil {
		testDB = setupTestDB()
	}
	return testDB
}
