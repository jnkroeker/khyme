/*
* Pro Tip: I want packages that PROVIDE, not CONTAIN
* If it doesnt make sense to have a file of the same name as the package,
* its a good indication that the package CONTAINs
 */
package handlers

// Package handlers contains the full set of handler functions and routes
// supported by the Tasker web api.

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/dimfeld/httptreemux/v5"
	"github.com/jnkroeker/khyme/app/services/tasker/handlers/debug/check"
	"github.com/jnkroeker/khyme/app/services/tasker/handlers/v1/test"
	"go.uber.org/zap"
)

// If you look at the http/pprof GoDoc, all these endpoints already bound to DefaultServerMux
// DebugStandardLibraryMux registers all the debug routes from the standard lilbrary
// into a new mux, bypassing the use of the DefaultServerMux.
// Using the DefaultServerMux would be a security risk since a dependency could inject a
// handler into our service without us knowing it.
func DebugStandardLibraryMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Register all the standard library debug endpoints.
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}

// extend DebugStandardLibraryMux by adding our own endpoints
func DebugMux(build string, log *zap.SugaredLogger) http.Handler {
	mux := DebugStandardLibraryMux()

	// Register debug endpoints
	check_handlers := check.Handlers{
		Build: build,
		Log:   log,
	}
	mux.HandleFunc("/debug/readiness", check_handlers.Readiness)
	mux.HandleFunc("/debug/liveness", check_handlers.Liveness)

	return mux
}

// contains all the mandatory systems required by handlers
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
}

// constructs an http.Handler with all application routes defined
func APIMux(cfg APIMuxConfig) *httptreemux.ContextMux {
	mux := httptreemux.NewContextMux()

	test_handlers := test.Handlers{
		Log: cfg.Log,
	}

	mux.Handle(http.MethodGet, "/test", test_handlers.Test)

	return mux
}
