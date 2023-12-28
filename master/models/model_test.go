package models

import (
	"GalaxyEmpireWeb/repositories/sqlite"
	"testing"
)

func TestAutoMigrate(t *testing.T) {
	db := sqlite.GetTestDB()
	AutoMigrate(db)
}
