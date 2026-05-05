package detailederror

import (
	"log/slog"
	"net/http"
)

// HTTPHandlerFunc is an HTTP handler that returns an error to its middleware.
type HTTPHandlerFunc func(http.ResponseWriter, *http.Request) error

// NewHTTPMiddleware creates middleware that logs errors returned by HTTPHandlerFunc.
//
// The middleware logs only non-nil returned errors using [slog.Logger.ErrorContext].
// Details attached to the error with [With] or [WithMany] are added as top-level
// structured log fields. If logger is nil, [slog.Default] is used.
//
// Panics are not recovered or logged by this middleware.
func NewHTTPMiddleware(logger *slog.Logger) func(HTTPHandlerFunc) http.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next HTTPHandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if err := next(w, r); err != nil {
				details := GetDetails(err)
				attrs := make([]any, 0, len(details))
				for k, v := range details {
					attrs = append(attrs, slog.String(k, v))
				}
				logger.ErrorContext(r.Context(), err.Error(), attrs...)
			}
		}
	}
}
