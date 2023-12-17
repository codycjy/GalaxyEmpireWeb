package redis

import (
	"os"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func ConnectRDB() {
	addr := "localhost:6379"
	if os.Getenv("REDIS_HOST") != "" {
		addr = os.Getenv("REDIS_HOST")
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

func GetRedisDB() *redis.Client {
	if rdb == nil {
		ConnectRDB()
	}
	return rdb
}

