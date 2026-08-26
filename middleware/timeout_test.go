package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func TestTimeout(t *testing.T) {
	t.Run("writes 504 when handler exceeds timeout without writing", func(t *testing.T) {
		h := middleware.Timeout(50 * time.Millisecond)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				select {
				case <-r.Context().Done():
					return
				case <-time.After(100 * time.Millisecond):
				}
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusGatewayTimeout {
			t.Errorf("expected status %d, got %d", http.StatusGatewayTimeout, rec.Code)
		}
		if body := rec.Body.String(); body == "" ||
			(body != http.StatusText(http.StatusGatewayTimeout) && !contains(body, "Gateway Timeout")) {
			t.Errorf("expected body to contain 'Gateway Timeout', got %q", body)
		}
	})

	t.Run("returns 200 when handler completes before timeout", func(t *testing.T) {
		h := middleware.Timeout(100 * time.Millisecond)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if got := rec.Body.String(); got != "ok" {
			t.Errorf("expected body %q, got %q", "ok", got)
		}
	})

	t.Run("does not attempt second WriteHeader when handler wrote then slept past timeout", func(t *testing.T) {
		h := middleware.Timeout(50 * time.Millisecond)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
				select {
				case <-r.Context().Done():
					return
				case <-time.After(100 * time.Millisecond):
				}
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if got := rec.Body.String(); got != "ok" {
			t.Errorf("expected body %q, got %q", "ok", got)
		}
	})
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ensure context import is referenced even if not used in helpers
var _ = context.Background
