package storage

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
	"sync"
)

var single sync.Once
var redisInstance *redis.Client

func NewRedis() *redis.Client {
	single.Do(func() {
		redisInstance = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
			Password: "",
			DB:       0,
		})
	})
	return redisInstance
}
