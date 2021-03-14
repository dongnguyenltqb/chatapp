package infra

import (
	"os"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"
)

var once sync.Once
var rdb *redis.Client

func GetRedis() *redis.Client {
	once.Do(func() {
		db, _ := strconv.Atoi(os.Getenv("redis_db"))
		rdb = redis.NewClient(&redis.Options{
			Addr:     os.Getenv("redis_addr"),
			Password: os.Getenv("redis_password"),
			DB:       db,
		})
	})
	return rdb
}
