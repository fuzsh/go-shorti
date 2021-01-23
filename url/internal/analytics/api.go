package analytics

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
func RegisterHandlers(r *routing.RouteGroup, service Service, logger log.Logger, authHandler routing.Handler) {
	res := resource{service, logger}

	r.Use(authHandler)
	r.Get("", res.analytics)

}

type resource struct {
	service Service
	logger  log.Logger
}

func (res resource) analytics(c *routing.Context) error {
	stats, err := res.service.Analytic(queries{
		Unique: c.Query("unique", "false"),
		Date:   c.Query("date", "monthly"),
		Mode:   c.Query("mode", "all"),
	}, c.Get("user_id").(int))
	if err != nil {
		return errors.BadRequest(err.Error())
	}
	return c.Write(stats)
}
