package api

import (
	"fmt"

	"github.com/savsgio/atreugo/v11"
)

// API is our email server api
type API struct {
	server *atreugo.Atreugo
}

// New returns api instance of our app
func New(port int) *API {
	if port == 0 {
		port = 8000 // Default port
	}

	api := &API{
		server: atreugo.New(atreugo.Config{
			Addr:             fmt.Sprintf(":%v", port),
			GracefulShutdown: true,
		}),
	}

	api.setRoutes()
	api.registerMiddlewares()

	return api
}

func (api *API) setRoutes() {
	api.server.Path("POST", "/api/v1/", sendEmailView)
}

func (api *API) registerMiddlewares() {
	api.server.UseBefore(checkParamsMiddleware)
}

// ListenAndServe starts api server on declared port
func (api *API) ListenAndServe() error {
	return api.server.ListenAndServe()
}
