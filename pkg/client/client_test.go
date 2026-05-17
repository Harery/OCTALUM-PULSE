package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestServer returns an httptest.Server with a router covering all client paths.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(HealthResponse{Status: "healthy", Version: "1.0.0", Timestamp: "now"})
	})
	mux.HandleFunc("/api/v1/doctor", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		_ = json.NewEncoder(w).Encode(OperationResponse{Success: true, Message: "ok"})
	})
	mux.HandleFunc("/api/v1/update", func(w http.ResponseWriter, r *http.Request) {
		// echo dry_run back as message for verification
		body, _ := io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(OperationResponse{Success: true, Message: string(body)})
	})
	mux.HandleFunc("/api/v1/cleanup", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(OperationResponse{Success: true, Message: "cleaned"})
	})
	mux.HandleFunc("/api/v1/history", func(w http.ResponseWriter, r *http.Request) {
		// reflect the limit query param back as a single operation's description
		limit := r.URL.Query().Get("limit")
		_ = json.NewEncoder(w).Encode(HistoryResponse{
			Operations: []OperationItem{{ID: 1, Type: "update", Description: limit, Status: "success", Timestamp: "t"}},
		})
	})

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return ts
}

func TestNewClientDefaults(t *testing.T) {
	c := NewClient("http://x")
	if c.baseURL != "http://x" {
		t.Errorf("baseURL = %q", c.baseURL)
	}
	if c.httpClient.Timeout != 30*time.Second {
		t.Errorf("default timeout = %v, want 30s", c.httpClient.Timeout)
	}
	if c.token != "" {
		t.Errorf("token = %q, want empty", c.token)
	}
}

func TestWithToken(t *testing.T) {
	c := NewClient("http://x", WithToken("secret"))
	if c.token != "secret" {
		t.Errorf("token = %q, want secret", c.token)
	}
}

func TestWithTimeout(t *testing.T) {
	c := NewClient("http://x", WithTimeout(2*time.Second))
	if c.httpClient.Timeout != 2*time.Second {
		t.Errorf("timeout = %v", c.httpClient.Timeout)
	}
}

func TestGetHealth(t *testing.T) {
	ts := newTestServer(t)
	c := NewClient(ts.URL)
	h, err := c.GetHealth(context.Background())
	if err != nil {
		t.Fatalf("GetHealth: %v", err)
	}
	if h.Status != "healthy" || h.Version != "1.0.0" {
		t.Errorf("response = %+v", h)
	}
}

func TestRunDoctor(t *testing.T) {
	ts := newTestServer(t)
	c := NewClient(ts.URL)
	res, err := c.RunDoctor(context.Background())
	if err != nil {
		t.Fatalf("RunDoctor: %v", err)
	}
	if !res.Success {
		t.Errorf("Success = false")
	}
}

func TestRunUpdate(t *testing.T) {
	ts := newTestServer(t)
	c := NewClient(ts.URL)
	opts := &UpdateOptions{DryRun: true, SecurityOnly: true, Exclude: []string{"foo"}}
	res, err := c.RunUpdate(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunUpdate: %v", err)
	}
	if !strings.Contains(res.Message, `"dry_run":true`) {
		t.Errorf("server did not see body: %q", res.Message)
	}
}

func TestRunCleanup(t *testing.T) {
	ts := newTestServer(t)
	c := NewClient(ts.URL)
	res, err := c.RunCleanup(context.Background(), &CleanupOptions{All: true})
	if err != nil {
		t.Fatalf("RunCleanup: %v", err)
	}
	if res.Message != "cleaned" {
		t.Errorf("Message = %q", res.Message)
	}
}

func TestGetHistory(t *testing.T) {
	ts := newTestServer(t)
	c := NewClient(ts.URL)
	hist, err := c.GetHistory(context.Background(), 25)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if len(hist.Operations) != 1 || hist.Operations[0].Description != "25" {
		t.Errorf("history = %+v", hist.Operations)
	}
}

func TestErrorStatusPropagates(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"boom"}`, http.StatusInternalServerError)
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	c := NewClient(ts.URL)
	_, err := c.GetHealth(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "API error") {
		t.Errorf("error message = %q", err.Error())
	}
}

func TestErrorOnAllEndpoints(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusBadRequest)
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	c := NewClient(ts.URL)
	ctx := context.Background()

	if _, err := c.RunDoctor(ctx); err == nil {
		t.Error("RunDoctor: want err")
	}
	if _, err := c.RunUpdate(ctx, &UpdateOptions{}); err == nil {
		t.Error("RunUpdate: want err")
	}
	if _, err := c.RunCleanup(ctx, &CleanupOptions{}); err == nil {
		t.Error("RunCleanup: want err")
	}
	if _, err := c.GetHistory(ctx, 5); err == nil {
		t.Error("GetHistory: want err")
	}
}

func TestInvalidJSONResponse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	c := NewClient(ts.URL)
	if _, err := c.GetHealth(context.Background()); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestContextCancellation(t *testing.T) {
	// Use a server that hangs until the client disconnects.
	done := make(chan struct{})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		close(done)
	}))
	t.Cleanup(ts.Close)

	c := NewClient(ts.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel
	_, err := c.GetHealth(ctx)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestBadBaseURL(t *testing.T) {
	c := NewClient("://bad")
	if _, err := c.GetHealth(context.Background()); err == nil {
		t.Fatal("expected request creation error")
	}
}

func TestRequestFailure(t *testing.T) {
	// Point at an unreachable address.
	c := NewClient("http://127.0.0.1:1")
	c.httpClient.Timeout = 100 * time.Millisecond
	if _, err := c.GetHealth(context.Background()); err == nil {
		t.Fatal("expected network error")
	}
}

func TestAuthHeaderSent(t *testing.T) {
	var gotAuth string
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	c := NewClient(ts.URL, WithToken("topsecret"))
	if _, err := c.GetHealth(context.Background()); err != nil {
		t.Fatalf("GetHealth: %v", err)
	}
	if gotAuth != "Bearer topsecret" {
		t.Errorf("Authorization = %q, want Bearer topsecret", gotAuth)
	}
}

func TestContentTypeHeaderSent(t *testing.T) {
	var gotCT string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/update", func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		_, _ = w.Write([]byte(`{}`))
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	c := NewClient(ts.URL)
	if _, err := c.RunUpdate(context.Background(), &UpdateOptions{}); err != nil {
		t.Fatalf("RunUpdate: %v", err)
	}
	if gotCT != "application/json" {
		t.Errorf("Content-Type = %q", gotCT)
	}
}
