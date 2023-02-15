package test

import (
	"context"
	"errors"
	"math/rand"
	"net/http"

	"github.com/jnkroeker/khyme/business/sys/validate"
	"github.com/jnkroeker/khyme/foundation/web"
	"go.uber.org/zap"
)

type Handlers struct {
	Log *zap.SugaredLogger
}

/*
 * Test transformed from an http.Handler (accepting w and r, without response) attached directly to httptreemux in APIMux()
 * into an instance of our custom Handler function type (as defined in foundation),
 * provided as the last parameter to our custom Handle() method on our App type (from foundation/web)
 */
// Context is critically important. Getting context out of the request early is critical to debugging and tracing.
// Only the handlers are going to be working with request context (its state and any timeouts),
// it wont be passed down thru the call stack.
func (h Handlers) Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	// test shutdown error handling
	if n := rand.Intn(100); n%2 == 0 {
		return validate.NewRequestError(errors.New("trusted error"), http.StatusBadRequest)
	}

	// the ok response is specific to each handler
	status := struct {
		Status string
	}{
		Status: "OK",
	}

	// I remove response marshalling protocol, error handling and logging
	// from the handler functions to ensure consistency of each of their
	// implementation across all handlers.
	// Handlers return an error, if there is one, with confidence that
	// there will be code to address it.

	// the foundational web package is setting policy for how we communicate
	// this will create consistency across all of our products
	return web.Respond(ctx, w, status, http.StatusOK)
}
