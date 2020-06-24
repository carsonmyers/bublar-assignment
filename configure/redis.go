package configure

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
)

// RedisConfig ENV configuration for redis connections
type RedisConfig struct {
	Host     string
	Port     uint
	Password string `json:"-"`
	DB       uint
}

// DefaultRedisConfig used if no config override is set
var DefaultRedisConfig = RedisConfig{
	Host:     "127.0.0.1",
	Port:     6379,
	Password: "",
	DB:       0,
}

func (c *RedisConfig) String() string {
	var b strings.Builder
	addParam(&b, "host", c.Host)
	addParam(&b, "port", c.Port)
	addParam(&b, "db", c.DB)

	return b.String()
}

// ConnectionOptions get redis connection options from the config
func (c *RedisConfig) ConnectionOptions() *redis.Options {
	port := strconv.FormatUint(uint64(c.Port), 10)
	return &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", c.Host, port),
		Password: c.Password,
		DB:       int(c.DB),
	}
}

var redisConfig *RedisConfig

// Redis set the config
func Redis(config *RedisConfig) {
	redisConfig = config
}

// GetRedis get the config
func GetRedis() *RedisConfig {
	if redisConfig == nil {
		return &DefaultRedisConfig
	}

	return redisConfig
}
