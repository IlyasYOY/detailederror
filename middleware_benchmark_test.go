package detailederror_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IlyasYOY/detailederror"
	"google.golang.org/grpc"
)

type discardHandler struct{}

func (discardHandler) Enabled(context.Context, slog.Level) bool { return true }

func (discardHandler) Handle(context.Context, slog.Record) error { return nil }

func (discardHandler) WithAttrs([]slog.Attr) slog.Handler { return discardHandler{} }

func (discardHandler) WithGroup(string) slog.Handler { return discardHandler{} }

func benchmarkDetailedError() error {
	return detailederror.WithMany(
		errors.New("benchmark error"),
		"user_id", "42",
		"account_id", "100",
		"operation", "create",
		"resource", "invoice",
		"region", "eu",
		"request_id", "req-1",
		"attempt", "2",
		"status", "failed",
	)
}

func BenchmarkHTTPMiddlewareErrorWithDetails(b *testing.B) {
	logger := slog.New(discardHandler{})
	middleware := detailederror.NewHTTPMiddleware(logger)
	detailedErr := benchmarkDetailedError()
	req := httptest.NewRequest(http.MethodGet, "/benchmark", nil)
	rec := httptest.NewRecorder()
	wrapped := middleware(func(http.ResponseWriter, *http.Request) error {
		return detailedErr
	})

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		wrapped(rec, req)
	}
}

func BenchmarkGRPCUnaryServerInterceptorErrorWithDetails(b *testing.B) {
	logger := slog.New(discardHandler{})
	interceptor := detailederror.NewGRPCUnaryServerInterceptor(logger)
	detailedErr := benchmarkDetailedError()
	info := &grpc.UnaryServerInfo{FullMethod: "/benchmark.Service/Method"}
	handler := func(context.Context, any) (any, error) {
		return nil, detailedErr
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_, _ = interceptor(context.Background(), "request", info, handler)
	}
}
