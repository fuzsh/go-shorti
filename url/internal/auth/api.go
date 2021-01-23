package auth

import (
	routing "github.com/go-ozzo/ozzo-routing/v2"
	"url/internal/errors"
	"url/pkg/log"
)

type Response struct {
	Message string `json:"message"`
}

var (
	//SuccessfulResponse for HttpStatusOK
	SuccessfulResponse = "Successful"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *routing.RouteGroup, service Service, logger log.Logger) {
	res := resource{service, logger}

	// routes related to auth with email and password
	r.Post("/signup", res.signUp)
	r.Post("/signin", res.signIn)
	r.Post("/email/verify-code", res.verifyEmail)

	// routes related to jwt token generation and expire them
	r.Post("/refresh-token", res.refresh)
}

type resource struct {
	service Service
	logger  log.Logger
}


func (res resource) signUp(c *routing.Context) error {
	input := AuthenticateWithEmailRequest{}
	if err := c.Read(&input); err != nil {
		res.logger.With(c.Request.Context()).Info("email", err)
		return errors.BadRequest("")
	}
	if err := res.service.SignUp(c.Request.Context(), input); err != nil {
		return err
	}
	return c.WriteWithStatus(Response{Message: SuccessfulResponse}, 201)
}

func (res resource) verifyEmail(c *routing.Context) error {
	input := VerifyEmailRequest{}
	if err := c.Read(&input); err != nil {
		res.logger.With(c.Request.Context()).Info("email", err)
		return errors.BadRequest("")
	}
	if err := res.service.VerifyEmail(c.Request.Context(), input); err != nil {
		return err
	}
	return c.Write(Response{Message: SuccessfulResponse})
}

func (res resource) signIn(c *routing.Context) error {
	input := AuthenticateWithEmailRequest{}
	if err := c.Read(&input); err != nil {
		res.logger.With(c.Request.Context()).Info("email", err)
		return errors.BadRequest("")
	}
	tokens, err := res.service.SignIn(c.Request.Context(), input)
	if err != nil {
		return err
	}
	return c.Write(tokens)
}

// refresh token
func (res resource) refresh(c *routing.Context) error {
	input := RefreshTokenRequest{}
	if err := c.Read(&input); err != nil {
		res.logger.With(c.Request.Context()).Info(err)
		return errors.BadRequest("")
	}
	tokens, err := res.service.Refresh(c.Request.Context(), input)
	if err != nil {
		return err
	}
	return c.Write(tokens)
}