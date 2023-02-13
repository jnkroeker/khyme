package test

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type Handlers struct {
	Log *zap.SugaredLogger
}

// Test transformed from an http.Handler (accepting w and r, without response) attached directly to httptreemux
// into an instance of our custom Handler function type (as defined in foundation),
// provided as the last parameter to our custom Handle() method on our App type (from foundation/web)
func (h Handlers) Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	status := struct {
		Status string
	}{
		Status: "OK",
	}

	statusCode := http.StatusOK

	if err := response(w, statusCode, status); err != nil {
		h.Log.Errorw("test", "ERROR", err)
		return err
	}

	h.Log.Infow("test", "statusCode", statusCode, "method", r.Method, "path", r.URL.Path, "remoteaddr", r.RemoteAddr)

	return json.NewEncoder(w).Encode(status)
}

func response(w http.ResponseWriter, statusCode int, data interface{}) error {
	// convert response to json
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")

	// Write the status code to the response.
	w.WriteHeader(statusCode)

	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}
