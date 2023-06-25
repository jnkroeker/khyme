package mid

import (
	"context"
	"errors"
	"net/http"

	"github.com/jnkroeker/khyme/business/web/auth"
	"github.com/jnkroeker/khyme/foundation/web"
)

// Set of error variables for handling user group errors
var (
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// Authenticate validates a JWT from the Authorization header
func Authenticate(a *auth.Auth) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			claims, err := a.Authenticate(ctx, r.Header.Get("authorization"))
			if err != nil {
				return auth.NewAuthError("authenticate: failed: %s", err)
			}

			ctx = auth.SetClaims(ctx, claims)

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
