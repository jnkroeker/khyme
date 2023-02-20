package mid

import (
	"context"
	"net/http"
	"time"

	"github.com/jnkroeker/khyme/foundation/web"
	"go.uber.org/zap"
)

// Defining a function that returns a Middleware allows for using
// the parameter passed into the function inside of the Middleware
// through the closure created by the function definition
func Logger(log *zap.SugaredLogger) web.Middleware {

	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			// help with logging by getting values from the request context
			v, err := web.GetValues(ctx)
			if err != nil {
				return err
			}

			log.Infow("request started", "traceid", v.TraceID, "method", r.Method, "path", r.URL.Path, "remoteaddr", r.RemoteAddr)

			err = handler(ctx, w, r)

			log.Infow("request completed", "traceid", v.TraceID, "method", r.Method, "path", r.URL.Path, "remoteaddr", r.RemoteAddr,
				"statuscode", v.StatusCode, "since", time.Since(v.Now))

			return err
		}

		return h
	}

	return m
}
