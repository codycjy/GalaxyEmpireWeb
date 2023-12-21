package main

import (
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/repositories/mysql"
	"GalaxyEmpireWeb/repositories/sqlite"
	"GalaxyEmpireWeb/routes"
	"GalaxyEmpireWeb/services/accountservice"
	"GalaxyEmpireWeb/services/userservice"
	"os"

	"gorm.io/gorm"
)

var services = make(map[string]interface{})

func servicesInit(db *gorm.DB) {
	userservice.InitService(db)
	accountservice.InitService(db)
}

func main() {
	var db *gorm.DB
	if os.Getenv("MODE") == "DEBUG" {
		db = sqlite.GetTestDB()
	} else {
		db = mysql.GetDB()
	}
	servicesInit(db)
	db.AutoMigrate(
		&models.User{},
		&models.Account{},
		&models.RouteTask{},
		&models.Fleet{})
	r := routes.RegisterRoutes(services)
	r.Run(":9333")
}
