package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"OCTALUM-PULSE/internal/config"
	"OCTALUM-PULSE/internal/platform"
	"OCTALUM-PULSE/internal/state"
)

func init() {
	if platform.DetectedPlatform == nil {
		platform.DetectedPlatform = &platform.Info{
			OS: "linux", Distribution: "ubuntu", Version: "22.04", VersionID: "22.04",
			PackageManager: "apt", InitSystem: "systemd", Arch: "amd64", Kernel: "test",
		}
	}
}

func newAgentWithState(t *testing.T) *Agent {
	t.Helper()
	cfg := config.DefaultConfig()
	dbPath := filepath.Join(t.TempDir(), "agent-test.db")
	st, err := state.New(dbPath)
	if err != nil {
		t.Fatalf("state.New: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	a := NewAgent(cfg)
	a.state = st
	return a
}

func TestNewAgent(t *testing.T) {
	cfg := config.DefaultConfig()
	a := NewAgent(cfg)
	if a == nil || a.config != cfg {
		t.Fatalf("NewAgent did not store config: %+v", a)
	}
	if a.state != nil || a.platform != nil || a.server != nil {
		t.Errorf("expected zero fields, got %+v", a)
	}
}

func TestHandleHealth(t *testing.T) {
	a := NewAgent(config.DefaultConfig())
	rec := httptest.NewRecorder()
	a.handleHealth(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v body=%q", err, rec.Body.String())
	}
	if got["status"] != "healthy" {
		t.Errorf("status field = %q", got["status"])
	}
	if !strings.Contains(rec.Body.String(), "timestamp") {
		t.Errorf("missing timestamp")
	}
}

func TestHandleReady(t *testing.T) {
	t.Run("no state", func(t *testing.T) {
		a := NewAgent(config.DefaultConfig())
		rec := httptest.NewRecorder()
		a.handleReady(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("status = %d, want 503", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "not ready") {
			t.Errorf("body = %q", rec.Body.String())
		}
	})
	t.Run("with state", func(t *testing.T) {
		a := newAgentWithState(t)
		rec := httptest.NewRecorder()
		a.handleReady(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"ready"`) {
			t.Errorf("body = %q", rec.Body.String())
		}
	})
}

func TestRunHealthCheck(t *testing.T) {
	// no-op when state is nil
	a := NewAgent(config.DefaultConfig())
	a.runHealthCheck(context.Background()) // must not panic

	// records metric when state present
	a = newAgentWithState(t)
	a.runHealthCheck(context.Background())
}

func TestShutdown(t *testing.T) {
	a := NewAgent(config.DefaultConfig())
	// Shutdown with no server should be safe
	if err := a.Shutdown(); err != nil {
		t.Errorf("Shutdown without server: %v", err)
	}

	// Shutdown with a real (unstarted) server
	a.server = &http.Server{}
	if err := a.Shutdown(); err != nil {
		t.Errorf("Shutdown with server: %v", err)
	}
}

func TestRunScheduledTasksRespectsContext(t *testing.T) {
	a := newAgentWithState(t)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		a.runScheduledTasks(ctx)
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("runScheduledTasks did not return after context cancellation")
	}
}

// TestStart is intentionally omitted: Agent.Start binds to a fixed :8080
// port and spawns background goroutines (health server, scheduled tasks)
// that may outlive the test goroutine, which the race detector flags.
// The individual building blocks (handlers, runHealthCheck, Shutdown,
// runScheduledTasks) are covered directly.
