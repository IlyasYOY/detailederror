package detailederror_test

import (
	"context"
	"errors"
	"testing"

	"github.com/IlyasYOY/detailederror"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"
)

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
