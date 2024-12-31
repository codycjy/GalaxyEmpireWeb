package main

import (
	"GalaxyEmpireWeb/config"
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
	"log"
	"os"

	r "github.com/redis/go-redis/v9"

	"go.uber.org/zap"
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
	if err := config.LoadPrices(""); err != nil {
		log.Fatal("Failed to load price configuration", zap.Error(err))
	}

	rdb = redis.GetRedisDB()
	mq = queue.GetRabbitMQ()

	db = mysql.GetDB()

	models.AutoMigrate(db)

	servicesInit(db, rdb, mq)

	fmt.Println("Server is running on port 9333")

	r := routes.RegisterRoutes()
	r.Run(":9333")

}
