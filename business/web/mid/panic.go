package mid

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/jnkroeker/khyme/business/sys/metrics"
	"github.com/jnkroeker/khyme/foundation/web"
)

// Recover from panics and turn the panic into an error
// to be reported in Metrics middleware and caught in Error middleware
func Panics() web.Middleware {

	m := func(handler web.Handler) web.Handler {

		// err is the name of the return argument for this handler function.
		// we use this technique in the defer function to return the error to be caught
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {

			defer func() {
				if rec := recover(); rec != nil {
					trace := debug.Stack()

					err = fmt.Errorf("PANIC [%v] TRACE [%s]", rec, string(trace))

					metrics.AddPanics(ctx)
				}
			}()

			// Call the next handler and set its return value in the error variable
			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
