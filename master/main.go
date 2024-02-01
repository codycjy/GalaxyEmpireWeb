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
var messageChan chan *queue.DelayedMessage
var taskChan chan *models.TaskItem
var responseChan chan []byte

func servicesInit(
	db *gorm.DB,
	rdb *r.Client) {
	userservice.InitService(db, rdb)
	accountservice.InitService(db, rdb)
	captchaservice.InitCaptchaService(rdb)
}

func queueInit(mq *queue.RabbitMQConnection) {

	taskChan = make(chan *models.TaskItem)
	responseChan = make(chan []byte)
	messageChan = make(chan *queue.DelayedMessage)

	mqProducer := queue.NewRabbitMQProducer(mq, messageChan)
	mqProducer.StartPublishing("") // WARN: exchange is empty
	// TODO: init service with queue
	taskservice.InitService(rdb, db, taskChan, messageChan, responseChan)

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
	servicesInit(db, rdb)

	r := routes.RegisterRoutes(services)
	r.Run(":9333")
}
