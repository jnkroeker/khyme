package web

import (
	"context"
	"encoding/json"
	"net/http"
)

func Respond(ctx context.Context, w http.ResponseWriter, data interface{}, statusCode int) error {

	// dont catch the error because this is debugging
	// the request needs to go thru regardless
	SetStatusCode(ctx, statusCode)

	// If there is no data to marshall, just set status code
	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	// Convert the response value to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Set content type and headers once marshaling has succeeded
	w.Header().Set("Content-Type", "application/json")

	// Write the status code to the response
	w.WriteHeader(statusCode)

	// Send the result back to the client
	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}
