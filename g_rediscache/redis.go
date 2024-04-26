package g_rediscache

import (
	"github.com/go-redis/redis/v8"
)

var redisClient *redis.Client

func InitRedisClient(cli *redis.Client) {
	redisClient = cli
}

func GetRedisClient() *redis.Client {
	return redisClient
}
