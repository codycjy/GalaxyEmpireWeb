package main

import (
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/queue"
	"GalaxyEmpireWeb/repositories/mysql"
	"GalaxyEmpireWeb/repositories/sqlite"
	"GalaxyEmpireWeb/routes"
	"GalaxyEmpireWeb/services/accountservice"
	"GalaxyEmpireWeb/services/taskservice"
	"GalaxyEmpireWeb/services/userservice"
	"os"

	"gorm.io/gorm"
)

var services = make(map[string]interface{})

func servicesInit(db *gorm.DB, mq *queue.RabbitMQConnection) {
	userservice.InitService(db)
	accountservice.InitService(db)
	taskservice.InitService(db, mq)
}

func main() {
	var db *gorm.DB
	var mq *queue.RabbitMQConnection
	mq = queue.GetRabbitMQ()
	if os.Getenv("env") == "test" {
		db = sqlite.GetTestDB()
	} else {
		db = mysql.GetDB()
	}
	servicesInit(db, mq)
	db.AutoMigrate(
		&models.User{},
		&models.Account{},
		&models.RouteTask{},
		&models.Fleet{},
		&models.TaskLog{},
	)
	r := routes.RegisterRoutes(services)
	r.Run(":9333")
}
