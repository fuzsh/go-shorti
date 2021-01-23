package redis

import (
	"fmt"
	redisClient "github.com/gomodule/redigo/redis"
	"time"
)

type Redis struct {
	Pool *redisClient.Pool
}

func New(host, port, password string) (*Redis, error) {
	pool := &redisClient.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redisClient.Conn, error) {
			return redisClient.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
		},
	}

	return &Redis{pool}, nil
}