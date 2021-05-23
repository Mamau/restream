package storage

import (
	"github.com/go-redis/redis/v8"
	"sync"
)

var single sync.Once
var redisInstance *redis.Client

func NewRedis() *redis.Client {
	single.Do(func() {
		redisInstance = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})
	})
	return redisInstance
}
