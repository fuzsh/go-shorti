package jwt

import (
	"time"
	"url/pkg/redis"
)

// AccessDetails holds the data of the authenticated user
// userId is the _id mongodb id
// access uuid s the key in the redisDb
type AccessDetails struct {
	AccessUUID string
	UserID     int
}

// TokenDetails holds all the needed data in the jwt
type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUUID   string
	RefreshUUID  string
	AtExpires    int64
	RtExpires    int64
}

// Options to define the instance of the jwt pkg.
type Options struct {
	RefreshTokenValidTime time.Duration
	AuthTokenValidTime    time.Duration
	AccessSecret          string
	RefreshSecret         string
	Debug                 bool
	Redis                 *redis.Redis
}
