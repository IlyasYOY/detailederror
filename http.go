package detailederror

import (
	"log/slog"
	"net/http"
)

type HTTPHandlerFunc func(http.ResponseWriter, *http.Request) error

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
