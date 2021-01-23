package auth

import (
	redisClient "github.com/gomodule/redigo/redis"
	"time"
	"url/pkg/log"
	"url/pkg/redis"
)

// Repository encapsulates the logic to access from the data source.
type Repository interface {
	SetVerifyCode(string, string) error
	GetVerifyCode(string) (string, error)
	DelVerifyCode(string) error
}

// repository persists in database
type repository struct {
	redis  *redis.Redis
	logger log.Logger
}

// NewRepository creates a new repository
func NewRepository(redis *redis.Redis, logger log.Logger) Repository {
	return repository{redis, logger}
}

// SaveVerifyCode saves a new user in the database.
func (r repository) SetVerifyCode(key, value string) error {
	conn := r.redis.Pool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", key, value)
	if err != nil {
		return err
	}
	conn.Do("EXPIRE", key, 5*time.Minute)

	return nil
}

// SaveVerifyCode saves a new user in the database.
func (r repository) DelVerifyCode(key string) error {
	conn := r.redis.Pool.Get()
	defer conn.Close()
	 _, err := conn.Do("DEL", key)
	return err
}

// SaveVerifyCode saves a new user in the database.
func (r repository) GetVerifyCode(key string) (string, error) {
	conn := r.redis.Pool.Get()
	defer conn.Close()

	return redisClient.String(conn.Do("GET", key))
}
