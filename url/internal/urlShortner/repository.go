package urlShortner

import (
	"context"
	"fmt"
	redisClient "github.com/gomodule/redigo/redis"
	"math/rand"
	"strconv"
	"url/pkg/base62"
	"url/pkg/log"
	"url/pkg/redis"
	"url/pkg/stringSuggestion"
)

// Repository encapsulates the logic to access from the data source.
type Repository interface {
	Create(ctx context.Context, URI string, similarTo string) (string, error)
	FindOne(ctx context.Context, code string) (string, error)
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

type RandomItem struct {
	Id  uint64 `json:"id" redis:"id"`
	URL string `json:"url" redis:"url"`
}

type SuggestedItem struct {
	Id  string `json:"id" redis:"id"`
	URL string `json:"url" redis:"url"`
}

func (r repository) Create(ctx context.Context, URI, similarTo string) (string, error) {
	conn := r.redis.Pool.Get()
	defer conn.Close()

	if similarTo != "" {
		for used := true; used; used = r.isSuggestedStrExists(similarTo) {
			similarTo = stringSuggestion.Suggest(similarTo, 2, 11)
		}
		shortLink := SuggestedItem{similarTo, URI}
		_, err := conn.Do("HMSET", redisClient.Args{"Shortener:" + similarTo}.AddFlat(shortLink)...)
		if err != nil {
			return "", err
		}
		return similarTo, nil
	}
	var id uint64
	for used := true; used; used = r.isIDUsed(id) {
		id = rand.Uint64()
	}
	shortLink := RandomItem{id, URI}
	_, err := conn.Do("HMSET", redisClient.Args{"Shortener:" + strconv.FormatUint(id, 10)}.AddFlat(shortLink)...)
	if err != nil {
		return "", err
	}
	return base62.Encode(id), nil
}

func (r repository) FindOne(ctx context.Context, code string) (string, error) {
	conn := r.redis.Pool.Get()
	defer conn.Close()

	urlString, err := redisClient.String(conn.Do("HGET", "Shortener:"+code, "url"))

	if err != nil {
		return "", err
	} else if len(urlString) == 0 {
		decodedId, err := base62.Decode(code)
		if err != nil {
			return "", err
		}
		urlString, err = redisClient.String(conn.Do("HGET", "Shortener:"+strconv.FormatUint(decodedId, 10), "url"))
		if err != nil {
			return "", err
		} else if len(urlString) == 0 {
			return "", fmt.Errorf("%s not found", code)
		}
	}
	return urlString, nil
}

func (r repository) isIDUsed(ID uint64) bool {
	conn := r.redis.Pool.Get()
	defer conn.Close()
	exists, err := redisClient.Bool(conn.Do("EXISTS", "Shortener:"+strconv.FormatUint(ID, 10)))
	if err != nil {
		return false
	}
	return exists
}

func (r repository) isSuggestedStrExists(ID string) bool {
	conn := r.redis.Pool.Get()
	defer conn.Close()
	exists, err := redisClient.Bool(conn.Do("EXISTS", "Shortener:"+ID))
	if err != nil {
		return false
	}
	return exists
}

func (r repository) isAvailable(id uint64) bool {
	conn := r.redis.Pool.Get()
	defer conn.Close()
	exists, err := redisClient.Bool(conn.Do("EXISTS", "Shortener:"+strconv.FormatUint(id, 10)))
	if err != nil {
		return false
	}
	return !exists
}
