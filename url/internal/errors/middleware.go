package errors

import (
	"fmt"
	routing "github.com/go-ozzo/ozzo-routing/v2"
	"github.com/go-playground/validator/v10"
	"github.com/lib/pq"
	"net/http"
	"runtime/debug"
	"url/pkg/log"
)

// Handler creates a middleware that handles panics and errors encountered during HTTP request processing.
func Handler(logger log.Logger) routing.Handler {
	return func(c *routing.Context) (err error) {
		defer func() {
			l := logger.With(c.Request.Context())
			if e := recover(); e != nil {
				var ok bool
				if err, ok = e.(error); !ok {
					err = fmt.Errorf("%v", e)
				}

				l.Errorf("recovered from panic (%v): %s", err, debug.Stack())
			}

			if err != nil {
				res := buildErrorResponse(err)
				if res.StatusCode() == http.StatusInternalServerError {
					l.Errorf("encountered internal server error: %v", err)
				}
				c.Response.WriteHeader(res.StatusCode())
				if err = c.Write(res); err != nil {
					l.Errorf("failed writing error response: %v", err)
				}
				c.Abort() // skip any pending handlers since an error has occurred
				err = nil // return nil because the error is already handled
			}
		}()
		return c.Next()
	}
}

// buildErrorResponse builds an error response from an error.
func buildErrorResponse(err error) ErrorResponse {
	fmt.Println(err)
	switch err.(type) {
	case validator.ValidationErrors:
		return InvalidInput(err.(validator.ValidationErrors))
	case ErrorResponse:
		return err.(ErrorResponse)
	case routing.HTTPError:
		switch err.(routing.HTTPError).StatusCode() {
		case http.StatusNotFound:
			return NotFound("")
		default:
			return ErrorResponse{
				Status:  err.(routing.HTTPError).StatusCode(),
				Message: err.Error(),
			}
		}
	}
	if err, ok := err.(*pq.Error); ok {
		return ErrorResponse{
			Status:  400,
			Message: err.Code.Name(),
			Details: err.Message,
		}
	}
	return InternalServerError("")
}
