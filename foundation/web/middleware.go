package web

// Middleware allows code to run before or after a Handler.
// It removes boilerplate not directly concerned with the Handler.
type Middleware func(Handler) Handler

func wrapMiddleware(mw []Middleware, handler Handler) Handler {

	// Looping backwards over mw slice ensures
	// the first Middleware in the slice
	// is the first to be executed by requests
	for i := len(mw) - 1; i >= 0; i-- {
		h := mw[i]
		if h != nil {
			handler = h(handler)
		}
	}

	return handler
}
