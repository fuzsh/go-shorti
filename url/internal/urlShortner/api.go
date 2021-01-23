package urlShortner

import (
	routing "github.com/go-ozzo/ozzo-routing/v2"
	"net/http"
	"url/internal/errors"
	"url/pkg/log"
)

var (
	//SuccessfulResponse for HttpStatusOK
	SuccessfulResponse = "Successful"
)

type resource struct {
	service Service
	logger  log.Logger
}

type Response struct {
	Message string `json:"data"`
}

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *routing.RouteGroup, service Service, logger log.Logger, authHandler routing.Handler) {
	res := resource{service, logger}
	r.Get("<shortLink>", res.redirect)

	r.Use(authHandler)
	r.Post("/api/v1/encode", res.encode)
}

func (res resource) encode(c *routing.Context) error {
	input := InputDTO{}
	if err := c.Read(&input); err != nil {
		res.logger.With(c.Request.Context()).Info(err)
		return errors.BadRequest("")
	}
	url, err := res.service.EnCode(c.Request.Context(), input, c.Get("user_id").(int))
	if err != nil {
		return err
	}
	return c.Write(Response{Message: url})
}

func (res resource) redirect(c *routing.Context) error {
	path := c.Param("shortLink")
	uri, err := res.service.Load(c.Request, path)
	if err != nil {
		return errors.NotFound(err.Error())
	}
	http.Redirect(c.Response, c.Request, uri, http.StatusMovedPermanently)
	return nil
}
