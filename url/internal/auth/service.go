package auth

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"url/internal/errors"
	"url/pkg/jwt"
	"url/pkg/log"
	"url/pkg/notification"
	"url/pkg/util"
	"url/pkg/validators"
)

// Service encapsulates use case logic.
type Service interface {
	SignUp(context.Context, AuthenticateWithEmailRequest) error
	VerifyEmail(context.Context, VerifyEmailRequest) error
	SignIn(context.Context, AuthenticateWithEmailRequest) (*Tokens, error)
	Refresh(context.Context, RefreshTokenRequest) (*Tokens, error)
}

type AuthenticateWithEmailRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// TODO -> validate length of code
type VerifyEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  int    `json:"code" validate:"required"`
}

type RefreshTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

type Tokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type service struct {
	repo   Repository
	store  Store
	logger log.Logger
	jwt    *jwt.Auth
}

// NewService creates a new service.
func NewService(store Store, repo Repository, logger log.Logger, jwt *jwt.Auth) Service {
	return service{repo, store, logger, jwt}
}

func (s service) SignUp(ctx context.Context, req AuthenticateWithEmailRequest) error {
	if ok, err := validators.Validate(req); !ok {
		return err
	}
	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return err
	}
	// create user
	if err := s.store.CreateUser(User{
		Username:   req.Email,
		Password:   string(hashedPassword),
		IsVerified: false,
	}); err != nil {
		return err
	}
	// generate the code
	code, err := util.GenerateVerificationCode(6)
	if err != nil {
		return err
	}
	// save the code in redis database
	if err := s.repo.SetVerifyCode(req.Email, code); err != nil {
		return err
	}

	fmt.Println(code)
	// send email to user
	return s.sendEmail(req.Email, code)
}


func (s service) VerifyEmail(ctx context.Context, req VerifyEmailRequest) error {
	if ok, err := validators.Validate(req); !ok {
		return err
	}
	// get the code from database
	code, err := s.repo.GetVerifyCode(req.Email)
	if err != nil {
		return err
	}
	// check code
	if strconv.Itoa(req.Code) != code {
		return errors.BadRequest("incorrect code or expired")
	}
	// delete data from redis
	s.repo.DelVerifyCode(req.Email)
	// update user and set user verify email field to true
	return s.store.VerifyEmail(nil, req.Email)
}

func (s service) SignIn(ctx context.Context, req AuthenticateWithEmailRequest) (*Tokens, error) {
	if ok, err := validators.Validate(req); !ok {
		return nil, err
	}
	user, err := s.authenticate(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	jwtTokens, err := s.generateJWT(user.ID)
	if err != nil {
		return nil, err
	}
	return jwtTokens, nil
}

func (s service) Refresh(_ context.Context, req RefreshTokenRequest) (*Tokens, error) {
	if ok, err := validators.Validate(req); !ok {
		return nil, err
	}
	ts, err := s.jwt.RefreshToken(req.Token)
	if err != nil {
		return nil, err
	}
	return &Tokens{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
	}, nil
}

//func (s service) Logout(req *http.Request) error {
//	metadata, err := s.jwt.ExtractTokenMetadata(req)
//	if err != nil {
//		return err
//	}
//	return s.jwt.DeleteTokens(metadata)
//}

// authenticate authenticates a user using username and password.
// If email and password are correct and user's email is verified, a user returned.
func (s service) authenticate(ctx context.Context, email, password string) (User, error) {
	user, err := s.store.FindOneByEmail(email)
	fmt.Println(user)
	// check user emails verified and error state
	if err != nil || !user.IsVerified {
		return user, errors.Unauthorized("invalid user")
	}
	// check the password
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil && user.IsVerified == true {
		return user, nil
	}
	return User{}, errors.Unauthorized("invalid user")
}

// generateJWT generates a JWT that encodes an identity.
func (s service) generateJWT(ID int,) (*Tokens, error) {
	ts, err := s.jwt.CreateToken(ID)
	if err != nil {
		return nil, err
	}
	if err := s.jwt.CreateAuth(ID, ts); err != nil {
		return nil, err
	}
	return &Tokens{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
	}, nil
}

func (s service) sendEmail(userEmail, code string) error {
	receiver := []string{userEmail}
	details := notification.EmailDetails{
		To:          receiver,
		Subject:     "verify",
		ContentType: "text/plain",
		Body:        code,
	}
	return notification.SendEmail(details)
}
