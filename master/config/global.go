package config

import "time"

var TASK_GENERATOR_INTERVAL = time.Second * 30
var TASK_QUEUE_NAME = "task_queue"
var RESULT_QUEUE_NAME = "result_queue"
var DELAYED_EXCHANGE_NAME = "delayed_exchange"
var TASK_DELAY = time.Second * 5
