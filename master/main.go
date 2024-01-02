package main

import (
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/queue"
	"GalaxyEmpireWeb/repositories/mysql"
	"GalaxyEmpireWeb/repositories/redis"
	"GalaxyEmpireWeb/repositories/sqlite"
	"GalaxyEmpireWeb/routes"
	"GalaxyEmpireWeb/services/accountservice"
	"GalaxyEmpireWeb/services/taskservice"
	"GalaxyEmpireWeb/services/userservice"
	"os"

	r "github.com/redis/go-redis/v9"

	"gorm.io/gorm"
)

var services = make(map[string]interface{})

func servicesInit(
	db *gorm.DB,
	mq *queue.RabbitMQConnection,
	rdb *r.Client) {
	userservice.InitService(db, rdb)
	accountservice.InitService(db, rdb)
	taskservice.InitService(db, mq)
}

var db *gorm.DB
var mq *queue.RabbitMQConnection
var rdb *r.Client

func main() {
	rdb = redis.GetRedisDB()
	mq = queue.GetRabbitMQ()

	if os.Getenv("env") == "test" {
		db = sqlite.GetTestDB()
	} else {
		db = mysql.GetDB()
	}

	models.AutoMigrate(db)
	servicesInit(db, mq, rdb)

	r := routes.RegisterRoutes(services)
	r.Run(":9333")
}
