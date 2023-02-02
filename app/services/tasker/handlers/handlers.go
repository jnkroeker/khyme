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
