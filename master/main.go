package main

import (
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/queue"
	"GalaxyEmpireWeb/repositories/mysql"
	"GalaxyEmpireWeb/repositories/redis"
	"GalaxyEmpireWeb/repositories/sqlite"
	"GalaxyEmpireWeb/routes"
	"GalaxyEmpireWeb/services/accountservice"
	"GalaxyEmpireWeb/services/captchaservice"
	"GalaxyEmpireWeb/services/casbinservice"
	"GalaxyEmpireWeb/services/taskservice"
	"GalaxyEmpireWeb/services/userservice"
	"fmt"
	"os"

	r "github.com/redis/go-redis/v9"

	"gorm.io/gorm"
)

func servicesInit(
	db *gorm.DB,
	mq *queue.RabbitMQConnection,
	rdb *r.Client) {
	captchaservice.InitCaptchaService(rdb)
	enforcer, err := casbinservice.NewCasbinEnforcer(db, "config/model.conf")
	if err != nil {
		panic(err)
	}
	userservice.InitService(db, rdb, enforcer)
	accountservice.InitService(db, rdb, enforcer)
	taskservice.InitService(db, mq, enforcer)
}

var db *gorm.DB
var mq *queue.RabbitMQConnection
var rdb *r.Client
var enforcer casbinservice.Enforcer //WARN: Remember to initialize this variable before using it.

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
	fmt.Println("Server is running on port 9333")

	r := routes.RegisterRoutes()
	r.Run(":9333")

}
