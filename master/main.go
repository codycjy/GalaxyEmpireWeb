package main

import (
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/queue"
	"GalaxyEmpireWeb/repositories/mysql"
	"GalaxyEmpireWeb/repositories/redis"
	"GalaxyEmpireWeb/routes"
	"GalaxyEmpireWeb/services/accountservice"
	"GalaxyEmpireWeb/services/captchaservice"
	"GalaxyEmpireWeb/services/casbinservice"
	"GalaxyEmpireWeb/services/paymentservice"
	"GalaxyEmpireWeb/services/taskservice"
	"GalaxyEmpireWeb/services/userservice"
	"fmt"

	r "github.com/redis/go-redis/v9"

	"gorm.io/gorm"
)

func servicesInit(db *gorm.DB, rdb *r.Client, mq *queue.RabbitMQConnection) {
	captchaservice.InitCaptchaService(rdb)
	enforcer, err := casbinservice.NewCasbinEnforcer(db, "config/model.conf")
	if err != nil {
		panic(err)
	}
	userservice.InitService(db, enforcer)
	accountservice.InitService(db, enforcer)
	taskservice.InitService(db, mq, enforcer)
	paymentservice.InitService(db)
}

var rdb *r.Client
var db *gorm.DB
var mq *queue.RabbitMQConnection
var enforcer casbinservice.Enforcer //WARN: Remember to initialize this variable before using it.

func main() {
	rdb = redis.GetRedisDB()
	mq = queue.GetRabbitMQ()

	db = mysql.GetDB()

	models.AutoMigrate(db)

	servicesInit(db, rdb, mq)

	fmt.Println("Server is running on port 9333")

	r := routes.RegisterRoutes()
	r.Run(":9333")

}
