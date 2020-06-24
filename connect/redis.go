package connect

import (
	"errors"

	"github.com/carsonmyers/bublar-assignment/configure"
	"github.com/go-redis/redis"
	"go.uber.org/zap"
)

var redisdb *redis.Client

// Redis attempts to connect to redis. Use GetRedisClient to retrieve the client once connection is successful
func Redis() (*redis.Client, error) {
	if redisdb != nil {
		_, err := redisdb.Ping().Result()
		if err == nil {
			return redisdb, nil
		}
	}

	config := configure.GetRedis()

	log.Info("Connecting to redis", zap.String("config", config.String()))

	redisdb = redis.NewClient(config.ConnectionOptions())
	_, err := redisdb.Ping().Result()
	if err != nil {
		log.Error("Connection to redis failed", zap.Error(err))
		return nil, errors.New("Redis connection failed")
	}

	return redisdb, nil
}
