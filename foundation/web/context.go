package web

import (
	"context"
	"errors"
	"time"
)

// keys are how values are placed in context
type ctxKey int

const key ctxKey = 1

// values related to every single request
// that are placed into the context
type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}

func GetValues(ctx context.Context) (*Values, error) {
	v, ok := ctx.Value(key).(*Values)

	if !ok {
		return nil, errors.New("web value missing from context")
	}

	return v, nil
}

func GetTraceId(ctx context.Context) string {
	v, ok := ctx.Value(key).(*Values)

	if !ok {
		return "0000-000-000-0000"
	}

	return v.TraceID
}

func SetStatusCode(ctx context.Context, statusCode int) error {
	v, ok := ctx.Value(key).(*Values)

	if !ok {
		return errors.New("web value missing from context")
	}

	v.StatusCode = statusCode
	return nil
}
