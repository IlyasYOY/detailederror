package detailederror

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
)

// NewGRPCUnaryServerInterceptor creates a unary server interceptor that logs handler errors.
//
// The interceptor logs only non-nil errors returned by the handler using
// [slog.Logger.ErrorContext]. Details attached to the error with [With] or
// [WithMany] are added as top-level structured log fields. If logger is nil,
// [slog.Default] is used.
//
// Panics are not recovered or logged by this interceptor.
func NewGRPCUnaryServerInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	if logger == nil {
		logger = slog.Default()
	}

	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			details := GetDetails(err)
			attrs := make([]any, 0, len(details))
			for k, v := range details {
				attrs = append(attrs, slog.String(k, v))
			}
			logger.ErrorContext(ctx, err.Error(), attrs...)
		}
		return resp, err
	}
}
