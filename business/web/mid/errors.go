package mid

import (
	"context"
	"net/http"

	"github.com/jnkroeker/khyme/business/sys/validate"
	"github.com/jnkroeker/khyme/foundation/web"
	"go.uber.org/zap"
)

func Errors(log *zap.SugaredLogger) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			v, err := web.GetValues(ctx)
			if err != nil {
				return web.NewShutdownError("web value missing from context")
			}

			if err = handler(ctx, w, r); err != nil {
				log.Errorw("ERROR", "traceid", v.TraceID, "ERROR", err)

				var er validate.ErrorResponse
				var status int

				// what root error type did we get back?
				// form an ErrorResponse and status depending on error type
				switch act := validate.Cause(err).(type) {
				// data model validation failed
				case validate.FieldErrors:
					er = validate.ErrorResponse{
						Error:  "data validation error",
						Fields: act.Error(),
					}
					status = http.StatusBadRequest
				case *validate.RequestError:
					er = validate.ErrorResponse{
						Error: act.Error(),
					}
					status = act.Status
				default:
					// untrusted error, return 500
					er = validate.ErrorResponse{
						Error: http.StatusText(http.StatusInternalServerError),
					}
					status = http.StatusInternalServerError
				}

				if err := web.Respond(ctx, w, er, status); err != nil {
					return err
				}

				if ok := web.IsShutdown(err); ok {
					return err
				}
			}

			return nil
		}

		return h
	}

	return m
}
