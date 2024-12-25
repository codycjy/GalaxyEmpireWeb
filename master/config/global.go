package config

import "time"

var TASK_GENERATOR_INTERVAL = time.Second * 30
var TASK_QUEUE_NAME = "task_queue"
var INSTANT_QUEUE_NAME = "instant_queue"
var RESULT_QUEUE_NAME = "result_queue"
var DELAYED_EXCHANGE_NAME = "delayed_exchange"
var TASK_DELAY = int64(5)           // seconds
var FAILED_TASK_DELAY = int64(3600) // seconds
var QUEUE_THRESHOLD = time.Minute * 60
