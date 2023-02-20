package web

import (
	"context"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/uuid"
)

// A Handler is a type that handles an http request
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type App struct {
	*httptreemux.ContextMux
	shutdown chan os.Signal
	mw       []Middleware
}

// ... spread operator allows zero to many Middleware functions
// This avoids requiring a user to pass a nil parameter (good design)
func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {
	return &App{
		ContextMux: httptreemux.NewContextMux(),
		shutdown:   shutdown,
		mw:         mw,
	}
}

// initiate a clean shutdown from inside the code if an integrity issue is discovered
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

func (a *App) Handle(method string, group string, path string, handler Handler, mw ...Middleware) {

	// Middlewares provided by parameter (thus specific to the path)
	// will be at the center of the onion, thus executed first
	handler = wrapMiddleware(mw, handler)

	// Application generic middleware wrapped into outermost layers
	handler = wrapMiddleware(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		v := Values{
			TraceID: uuid.New().String(),
			Now:     time.Now(),
		}

		ctx = context.WithValue(ctx, key, &v)

		// If an error gets here, past the error handler to the outer layer of the onion
		// life is really bad, so shutdown
		if err := handler(ctx, w, r); err != nil {
			a.SignalShutdown()
			return
		}
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}

	a.ContextMux.Handle(method, finalPath, h)
}
