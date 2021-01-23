package jwt

import (
	"net/http"

	routing "github.com/go-ozzo/ozzo-routing/v2"
)

// Handler returns a middleware that checks the access token.
func (a *Auth) Handler() routing.Handler {
	return func(c *routing.Context) error {
		//Extract the access token metadata
		metadata, err := a.ExtractTokenMetadata(c.Request)
		if err != nil {
			return routing.NewHTTPError(http.StatusUnauthorized)
		}
		userid, err := a.FetchAuth(metadata)
		if err != nil {
			return routing.NewHTTPError(http.StatusUnauthorized)
		}
		c.Set("user_id", userid)
		return c.Next()
	}
}
