package models

import (
	"GalaxyEmpireWeb/repositories/sqlite"
	"testing"
)

func TestAutoMigrate(t *testing.T) {
	db := sqlite.GetTestDB()

	models := []interface{}{
		&User{},
		&Account{},
		&RouteTask{},
		&DailyTask{},
		&ExtraTask{},
		&Fleet{},
		&TaskLog{},
	}

	if err := db.AutoMigrate(models...); err != nil {
		t.Errorf("AutoMigrate error: %v", err)
	}
}
