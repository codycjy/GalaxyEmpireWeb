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
	"GalaxyEmpireWeb/services/taskservice"
	"GalaxyEmpireWeb/services/userservice"
	"os"

	r "github.com/redis/go-redis/v9"

	"gorm.io/gorm"
)

var services = make(map[string]interface{})
var db *gorm.DB
var mq *queue.RabbitMQConnection
var rdb *r.Client

func servicesInit(
	db *gorm.DB,
	mq *queue.RabbitMQConnection,
	rdb *r.Client) {
	userservice.InitService(db, rdb)
	accountservice.InitService(db, rdb)
	taskservice.InitService(db, mq)
	captchaservice.InitCaptchaService(rdb)
}
func queueInit(mq *queue.RabbitMQConnection) {

	taskChan := make(chan queue.DelayedMessage)

	mqProducer := queue.NewRabbitMQProducer(mq, taskChan)
	mqProducer.StartPublishing("")
	// TODO: init service with queue

	messageChan := make(chan []byte)

	mqConsumer := queue.NewRabbitMQConsumer(mq, messageChan)
	taskHandler := taskservice.NewTaskHandler(rdb, db, messageChan)

	mqConsumer.StartConsuming("response")
	taskHandler.HandleResponse()

}

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
