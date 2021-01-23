package jwt

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"strconv"
	"strings"
	"time"
	redisClient "github.com/gomodule/redigo/redis"

	"github.com/google/uuid"
)

const (
	defaultRefreshTokenValidTime = time.Hour * 24 * 7
	defaultAuthTokenValidTime    = time.Minute * 15
)

// Auth is the model for jwt instance
type Auth struct {
	opts Options
}

// New create the instance for jwt and returns auth model.
func New(opts Options) (*Auth, error) {
	o := opts

	if o.AuthTokenValidTime <= 0 {
		o.AuthTokenValidTime = defaultAuthTokenValidTime
	}

	if o.RefreshTokenValidTime <= 0 {
		o.RefreshTokenValidTime = defaultRefreshTokenValidTime
	}

	if o.RefreshSecret == "" || o.AccessSecret == "" {
		return nil, errors.New("you should provide the access and refresh secret keys")
	}

	if o.Redis == nil {
		return nil, errors.New("you should initialize the redis client")
	}

	auth := &Auth{
		opts: o,
	}

	return auth, nil
}

// CreateToken generates the refresh and auth token.
func (a *Auth) CreateToken(userID int) (*TokenDetails, error) {
	var err error
	td := &TokenDetails{}

	td.AtExpires = time.Now().Add(a.opts.AuthTokenValidTime).Unix()
	td.AccessUUID = uuid.New().String()

	td.RtExpires = time.Now().Add(a.opts.RefreshTokenValidTime).Unix()
	td.RefreshUUID = td.AccessUUID + "++" + strconv.Itoa(userID)

	now := time.Now().Unix()

	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUUID
	atClaims["user_id"] = strconv.Itoa(userID)
	atClaims["exp"] = td.AtExpires
	atClaims["iat"] = now
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(a.opts.AccessSecret))
	if err != nil {
		return nil, err
	}

	//Creating Refresh Token
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUUID
	rtClaims["user_id"] = strconv.Itoa(userID)
	rtClaims["exp"] = td.RtExpires
	rtClaims["iat"] = now
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(a.opts.RefreshSecret))
	if err != nil {
		return nil, err
	}
	return td, nil
}

// CreateAuth save the credentials of jwt in the redisDb.
func (a *Auth) CreateAuth(userID int, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	conn := a.opts.Redis.Pool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", td.AccessUUID, strconv.Itoa(userID))
	if err != nil {
		return err
	}
	_, err = conn.Do("SET", td.RefreshUUID, strconv.Itoa(userID))
	if err != nil {
		return err
	}
	conn.Do("EXPIRE", td.AccessUUID, at.Sub(now))
	conn.Do("EXPIRE", td.RefreshUUID, rt.Sub(now))

	return nil
}

// verifyToken Parse, validate, and return a token.
// keyFunc will receive the parsed token and should return the key for validating.
func verifyToken(tokenString, secret string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

// get token from the request and then verify it
func (a *Auth) verifyTokenFromRequest(r *http.Request) (*jwt.Token, error) {
	tokenString := extractToken(r)
	return verifyToken(tokenString, a.opts.AccessSecret)
}

// ExtractTokenMetadata extract the token from the header Authorization
// verify token and if token is valid return token metadata
func (a *Auth) ExtractTokenMetadata(r *http.Request) (*AccessDetails, error) {
	token, err := a.verifyTokenFromRequest(r)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUUID, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, fmt.Errorf("access_uuid is not valid")
		}
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return nil, fmt.Errorf("user_id is not valid")
		}
		userID, _ := strconv.Atoi(userIDStr)
		accessDetails := &AccessDetails{
			AccessUUID: accessUUID,
			UserID:     userID,
		}
		conn := a.opts.Redis.Pool.Get()
		defer conn.Close()

		value, _ := redisClient.String(conn.Do("GET", claims["user_id"].(string)))

		valueArr := strings.Split(value, "_")
		if valueArr[0] == "change-password" && fmt.Sprintf("%f", claims["iat"].(float64)) < valueArr[1] {
			a.DeleteTokens(accessDetails)
			return nil, errors.New("user password change")
		}
		return accessDetails, nil
	}
	return nil, err
}

// FetchAuth get the the userId from the redisDb
func (a *Auth) FetchAuth(authD *AccessDetails) (int, error) {
	conn := a.opts.Redis.Pool.Get()
	defer conn.Close()
	userid, err := redisClient.String(conn.Do("GET", authD.AccessUUID))
	if err != nil {
		return 0, err
	} else if len(userid) == 0 {
		return 0, fmt.Errorf("%s not found", authD.AccessUUID)
	}
	userID, err := strconv.Atoi(userid)
	if err != nil || authD.UserID != userID {
		return 0, errors.New("unauthorized")
	}
	return userID, nil
}

// DeleteAuth deletes the provided uuid from redisDb
func (a *Auth) deleteAuth(givenUUID string) (int64, error) {
	conn := a.opts.Redis.Pool.Get()
	defer conn.Close()
	deleted, err := redisClient.Int64(conn.Do("DEL", givenUUID))
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

// DeleteTokens deletes the tokens base on AccessDetails from redisDb
func (a *Auth) DeleteTokens(authD *AccessDetails) error {
	conn := a.opts.Redis.Pool.Get()
	defer conn.Close()
	//get the refresh uuid
	refreshUUID := fmt.Sprintf("%s++%d", authD.AccessUUID, authD.UserID)
	//delete access token

	deletedAt, err := redisClient.Int64(conn.Do("DEL", authD.AccessUUID))
	if err != nil {
		return err
	}
	//delete refresh token
	deletedRt, err := redisClient.Int64(conn.Do("DEL", refreshUUID))
	if err != nil {
		return err
	}
	//When the record is deleted, the return value is 1
	if deletedAt != 1 || deletedRt != 1 {
		return errors.New("something went wrong")
	}
	return nil
}

// RefreshToken get the user refresh token and generate new token pairs if refreshToken is valid
func (a *Auth) RefreshToken(refreshToken string) (*TokenDetails, error) {
	token, err := verifyToken(refreshToken, a.opts.RefreshSecret)
	if err != nil {
		return nil, err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return nil, errors.New("unauthorized")
	}
	//Since token is valid, get the uuid:
	claims, ok := token.Claims.(jwt.MapClaims) //the token claims should conform to MapClaims
	if ok && token.Valid {
		refreshUUID, ok := claims["refresh_uuid"].(string) //convert the interface to string
		if !ok {
			return nil, errors.New("unauthorized")
		}

		conn := a.opts.Redis.Pool.Get()
		defer conn.Close()

		value, _ := redisClient.String(conn.Do("GET", claims["user_id"].(string)))
		valueArr := strings.Split(value, "_")
		if valueArr[0] == "change-password" && fmt.Sprintf("%f", claims["iat"].(float64)) < valueArr[1] {
			return nil, errors.New("user password change")
		}
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return nil, fmt.Errorf("user_id is not valid")
		}
		userID, _ := strconv.Atoi(userIDStr)

		//Delete the previous Refresh Token
		deleted, delErr := a.deleteAuth(refreshUUID)
		if delErr != nil || deleted == 0 { //if any goes wrong
			return nil, errors.New("unauthorized")
		}
		//Create new pairs of refresh and access tokens
		ts, createErr := a.CreateToken(userID)
		if createErr != nil {
			return nil, errors.New("unauthorized")
		}
		//save the tokens metadata to redisDb
		if err := a.CreateAuth(userID, ts); err != nil {
			return nil, errors.New("unauthorized")
		}
		return ts, nil
	}
	return nil, errors.New("unauthorized")
}

// validate token
func (a *Auth) tokenValid(r *http.Request) error {
	token, err := a.verifyTokenFromRequest(r)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok || !token.Valid {
		return err
	}
	return nil
}

// get the token from header
func extractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}
