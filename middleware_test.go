package detailederror_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"

	"google.golang.org/grpc"

	"github.com/IlyasYOY/detailederror"
	"github.com/google/go-cmp/cmp"
)

type capturedHandler struct {
	records []capturedRecord
}

type capturedRecord struct {
	msg   string
	attrs map[string]string
}

func (c *capturedHandler) Enabled(context.Context, slog.Level) bool { return true }

func (c *capturedHandler) Handle(_ context.Context, record slog.Record) error {
	recordAttrs := make(map[string]string)
	record.Attrs(func(attr slog.Attr) bool {
		recordAttrs[attr.Key] = attr.Value.String()
		return true
	})
	c.records = append(c.records, capturedRecord{
		msg:   record.Message,
		attrs: recordAttrs,
	})
	return nil
}

func (c *capturedHandler) WithAttrs(_ []slog.Attr) slog.Handler { return c }

func (c *capturedHandler) WithGroup(_ string) slog.Handler { return c }

func newTestLogger() (*slog.Logger, *capturedHandler) {
	ch := &capturedHandler{}
	return slog.New(ch), ch
}

func TestNewHTTPMiddleware_LogsErrorWithoutDetails(t *testing.T) {
	logger, handler := newTestLogger()
	middleware := detailederror.NewHTTPMiddleware(logger)
	err := errors.New("http error")
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	wrapped := middleware(func(_ http.ResponseWriter, _ *http.Request) error {
		return err
	})

	wrapped(rec, req)

	if len(handler.records) != 1 {
		t.Fatalf("records count = %d, want 1", len(handler.records))
	}
	if handler.records[0].msg != err.Error() {
		t.Fatalf("message = %q, want %q", handler.records[0].msg, err.Error())
	}
	if len(handler.records[0].attrs) != 0 {
		t.Fatalf("attrs = %v, want empty", handler.records[0].attrs)
	}
}

func TestNewHTTPMiddleware_LogsOneDetail(t *testing.T) {
	logger, handler := newTestLogger()
	middleware := detailederror.NewHTTPMiddleware(logger)
	err := errors.New("http error")
	detailedErr := detailederror.With(err, "user", "ilya")
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	wrapped := middleware(func(_ http.ResponseWriter, _ *http.Request) error {
		return detailedErr
	})

	wrapped(rec, req)

	if len(handler.records) != 1 {
		t.Fatalf("records count = %d, want 1", len(handler.records))
	}
	if diff := cmp.Diff(map[string]string{"user": "ilya"}, handler.records[0].attrs); diff != "" {
		t.Fatalf("attrs mismatch:\n%s", diff)
	}
}

func TestNewHTTPMiddleware_LogsManyDetails(t *testing.T) {
	logger, handler := newTestLogger()
	middleware := detailederror.NewHTTPMiddleware(logger)
	err := errors.New("http error")
	detailedErr := detailederror.WithMany(
		err,
		"user1", "ilya1",
		"user2", "ilya2",
	)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	wrapped := middleware(func(_ http.ResponseWriter, _ *http.Request) error {
		return detailedErr
	})

	wrapped(rec, req)

	if len(handler.records) != 1 {
		t.Fatalf("records count = %d, want 1", len(handler.records))
	}
	if diff := cmp.Diff(map[string]string{
		"user1": "ilya1",
		"user2": "ilya2",
	}, handler.records[0].attrs); diff != "" {
		t.Fatalf("attrs mismatch:\n%s", diff)
	}
}

func TestNewHTTPMiddleware_DoesNotLogOnNil(t *testing.T) {
	logger, handler := newTestLogger()
	middleware := detailederror.NewHTTPMiddleware(logger)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	wrapped := middleware(func(_ http.ResponseWriter, _ *http.Request) error { return nil })

	wrapped(rec, req)

	if len(handler.records) != 0 {
		t.Fatalf("records count = %d, want 0", len(handler.records))
	}
}

func TestNewHTTPMiddleware_DoesNotLogPanic(t *testing.T) {
	logger, handler := newTestLogger()
	middleware := detailederror.NewHTTPMiddleware(logger)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	wrapped := middleware(func(_ http.ResponseWriter, _ *http.Request) error {
		panic("boom")
	})

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("panic was not propagated")
		}
		if len(handler.records) != 0 {
			t.Fatalf("records count = %d, want 0", len(handler.records))
		}
	}()
	wrapped(rec, req)
}

func TestNewGRPCUnaryServerInterceptor_LogsErrorWithoutDetails(t *testing.T) {
	logger, handler := newTestLogger()
	interceptor := detailederror.NewGRPCUnaryServerInterceptor(logger)
	err := errors.New("grpc error")
	_, gotErr := interceptor(context.Background(), "request", &grpc.UnaryServerInfo{}, func(_ context.Context, _ any) (any, error) {
		return nil, err
	})

	if gotErr != err {
		t.Fatalf("error = %v, want %v", gotErr, err)
	}
	if len(handler.records) != 1 {
		t.Fatalf("records count = %d, want 1", len(handler.records))
	}
	if handler.records[0].msg != err.Error() {
		t.Fatalf("message = %q, want %q", handler.records[0].msg, err.Error())
	}
	if len(handler.records[0].attrs) != 0 {
		t.Fatalf("attrs = %v, want empty", handler.records[0].attrs)
	}
}

func TestNewGRPCUnaryServerInterceptor_LogsOneDetail(t *testing.T) {
	logger, handler := newTestLogger()
	interceptor := detailederror.NewGRPCUnaryServerInterceptor(logger)
	err := errors.New("grpc error")
	detailedErr := detailederror.With(err, "user", "ilya")
	_, gotErr := interceptor(context.Background(), "request", &grpc.UnaryServerInfo{}, func(_ context.Context, _ any) (any, error) {
		return nil, detailedErr
	})

	if gotErr != detailedErr {
		t.Fatalf("error = %v, want %v", gotErr, detailedErr)
	}
	if len(handler.records) != 1 {
		t.Fatalf("records count = %d, want 1", len(handler.records))
	}
	if diff := cmp.Diff(map[string]string{"user": "ilya"}, handler.records[0].attrs); diff != "" {
		t.Fatalf("attrs mismatch:\n%s", diff)
	}
}

func TestNewGRPCUnaryServerInterceptor_LogsManyDetails(t *testing.T) {
	logger, handler := newTestLogger()
	interceptor := detailederror.NewGRPCUnaryServerInterceptor(logger)
	err := errors.New("grpc error")
	detailedErr := detailederror.WithMany(
		err,
		"user1", "ilya1",
		"user2", "ilya2",
	)
	_, gotErr := interceptor(context.Background(), "request", &grpc.UnaryServerInfo{}, func(_ context.Context, _ any) (any, error) {
		return nil, detailedErr
	})

	if gotErr != detailedErr {
		t.Fatalf("error = %v, want %v", gotErr, detailedErr)
	}
	if len(handler.records) != 1 {
		t.Fatalf("records count = %d, want 1", len(handler.records))
	}
	if diff := cmp.Diff(map[string]string{
		"user1": "ilya1",
		"user2": "ilya2",
	}, handler.records[0].attrs); diff != "" {
		t.Fatalf("attrs mismatch:\n%s", diff)
	}
}

func TestNewGRPCUnaryServerInterceptor_DoesNotLogOnNil(t *testing.T) {
	logger, handler := newTestLogger()
	interceptor := detailederror.NewGRPCUnaryServerInterceptor(logger)
	_, gotErr := interceptor(context.Background(), "request", &grpc.UnaryServerInfo{}, func(_ context.Context, _ any) (any, error) {
		return "ok", nil
	})

	if gotErr != nil {
		t.Fatalf("error = %v, want nil", gotErr)
	}
	if len(handler.records) != 0 {
		t.Fatalf("records count = %d, want 0", len(handler.records))
	}
}

func TestNewGRPCUnaryServerInterceptor_DoesNotLogPanic(t *testing.T) {
	logger, handler := newTestLogger()
	interceptor := detailederror.NewGRPCUnaryServerInterceptor(logger)
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("panic was not propagated")
		}
		if len(handler.records) != 0 {
			t.Fatalf("records count = %d, want 0", len(handler.records))
		}
	}()
	_, _ = interceptor(context.Background(), "request", &grpc.UnaryServerInfo{}, func(_ context.Context, _ any) (any, error) {
		panic("boom")
	})
}
