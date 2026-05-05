package detailederror_test

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IlyasYOY/detailederror"
	"github.com/google/go-cmp/cmp"
)

type capturedHandler struct {
	records []capturedRecord
}

type capturedRecord struct {
	attrs map[string]string
	msg   string
}

func (c *capturedHandler) Write(p []byte) (int, error) {
	var data map[string]string
	if err := json.Unmarshal(p, &data); err != nil {
		return 0, err
	}
	recordAttrs := make(map[string]string)
	for k, v := range data {
		if k == "time" || k == "level" || k == "msg" {
			continue
		}
		recordAttrs[k] = v
	}
	c.records = append(c.records, capturedRecord{
		msg:   data["msg"],
		attrs: recordAttrs,
	})
	return len(p), nil
}

func newTestLogger() (*slog.Logger, *capturedHandler) {
	ch := &capturedHandler{}
	return slog.New(slog.NewJSONHandler(ch, nil)), ch
}

func TestNewHTTPMiddleware_LogsErrorWithoutDetails(t *testing.T) {
	logger, handler := newTestLogger()
	middleware := detailederror.NewHTTPMiddleware(logger)
	err := errors.New("http error")
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
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
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
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
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
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
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
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
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
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
