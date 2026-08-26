package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
)

func TestHeartbeat(t *testing.T) {
	tt := []struct {
		name        string
		method      string
		path        string
		wantStatus  int
		wantHeaders map[string]string
		wantBody    string
		nextCalled  bool
	}{
		{
			name:        "GET on heartbeat endpoint",
			method:      http.MethodGet,
			path:        "/ping",
			wantStatus:  http.StatusOK,
			wantHeaders: map[string]string{"Content-Type": "text/plain"},
			wantBody:    ".",
			nextCalled:  false,
		},
		{
			name:        "HEAD on heartbeat endpoint",
			method:      http.MethodHead,
			path:        "/ping",
			wantStatus:  http.StatusOK,
			wantHeaders: map[string]string{"Content-Type": "text/plain"},
			wantBody:    ".",
			nextCalled:  false,
		},
		{
			name:        "POST on heartbeat endpoint returns 405",
			method:      http.MethodPost,
			path:        "/ping",
			wantStatus:  http.StatusMethodNotAllowed,
			wantHeaders: map[string]string{"Allow": "GET, HEAD"},
			wantBody:    "",
			nextCalled:  false,
		},
		{
			name:        "PUT on heartbeat endpoint returns 405",
			method:      http.MethodPut,
			path:        "/ping",
			wantStatus:  http.StatusMethodNotAllowed,
			wantHeaders: map[string]string{"Allow": "GET, HEAD"},
			wantBody:    "",
			nextCalled:  false,
		},
		{
			name:        "DELETE on heartbeat endpoint returns 405",
			method:      http.MethodDelete,
			path:        "/ping",
			wantStatus:  http.StatusMethodNotAllowed,
			wantHeaders: map[string]string{"Allow": "GET, HEAD"},
			wantBody:    "",
			nextCalled:  false,
		},
		{
			name:        "GET on different path forwards to next",
			method:      http.MethodGet,
			path:        "/other",
			wantStatus:  http.StatusOK,
			wantHeaders: map[string]string{"X-Next": "yes"},
			wantBody:    "next",
			nextCalled:  true,
		},
		{
			name:        "POST on different path forwards to next",
			method:      http.MethodPost,
			path:        "/other",
			wantStatus:  http.StatusOK,
			wantHeaders: map[string]string{"X-Next": "yes"},
			wantBody:    "next",
			nextCalled:  true,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			invoked := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				invoked = true
				w.Header().Set("X-Next", "yes")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("next"))
			})

			h := middleware.Heartbeat("/ping")(next)

			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", rr.Code, tc.wantStatus)
			}
			for k, v := range tc.wantHeaders {
				if got := rr.Header().Get(k); got != v {
					t.Errorf("header %s: got %q, want %q", k, got, v)
				}
			}
			if body := rr.Body.String(); body != tc.wantBody {
				t.Errorf("body: got %q, want %q", body, tc.wantBody)
			}
			if invoked != tc.nextCalled {
				t.Errorf("next handler invoked: got %v, want %v", invoked, tc.nextCalled)
			}
		})
	}
}
