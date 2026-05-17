package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"OCTALUM-PULSE/internal/config"
	"OCTALUM-PULSE/internal/platform"
	"OCTALUM-PULSE/internal/state"
)

func init() {
	// Pre-seed platform detection so NewServer works on non-Linux test runners.
	if platform.DetectedPlatform == nil {
		platform.DetectedPlatform = &platform.Info{
			OS:             "linux",
			Distribution:   "ubuntu",
			Version:        "22.04",
			VersionID:      "22.04",
			PackageManager: "apt",
			InitSystem:     "systemd",
			Arch:           "amd64",
			Kernel:         "test",
		}
	}
}

// newTestServer constructs a Server with a fresh on-disk SQLite state for each test.
func newTestServer(t *testing.T) *Server {
	t.Helper()

	cfg := config.DefaultConfig()
	dbPath := filepath.Join(t.TempDir(), "pulse-test.db")
	st, err := state.New(dbPath)
	if err != nil {
		t.Fatalf("state.New() error = %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	srv, err := NewServer(cfg, st)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	return srv
}

// newTestServerNoState constructs a Server with state == nil to exercise nil-state code paths.
func newTestServerNoState(t *testing.T) *Server {
	t.Helper()
	cfg := config.DefaultConfig()
	srv, err := NewServer(cfg, nil)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	return srv
}

func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder, dst interface{}) {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), dst); err != nil {
		t.Fatalf("json.Unmarshal: %v (body=%q)", err, rec.Body.String())
	}
}

func TestHandleHealth(t *testing.T) {
	srv := newTestServer(t)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	decodeJSON(t, rec, &resp)
	if resp["status"] != "healthy" {
		t.Errorf("status field = %v, want healthy", resp["status"])
	}
	if _, ok := resp["timestamp"]; !ok {
		t.Errorf("missing timestamp")
	}
	if _, ok := resp["version"]; !ok {
		t.Errorf("missing version")
	}
}

func TestHandleReady(t *testing.T) {
	tests := []struct {
		name    string
		srv     func(*testing.T) *Server
		want    int
		ready   bool
	}{
		{"with state", newTestServer, http.StatusOK, true},
		{"without state", newTestServerNoState, http.StatusServiceUnavailable, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := tc.srv(t)
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))

			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}
			var resp map[string]interface{}
			decodeJSON(t, rec, &resp)
			if resp["ready"] != tc.ready {
				t.Errorf("ready = %v, want %v", resp["ready"], tc.ready)
			}
		})
	}
}

func TestHandleVersion(t *testing.T) {
	srv := newTestServer(t)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/version", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	decodeJSON(t, rec, &resp)
	if len(resp) == 0 {
		t.Errorf("empty version response")
	}
}

func TestHandlePlatform(t *testing.T) {
	srv := newTestServer(t)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/platform", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	decodeJSON(t, rec, &resp)
	for _, k := range []string{"os", "arch", "family", "supported"} {
		if _, ok := resp[k]; !ok {
			t.Errorf("missing key %q in platform response", k)
		}
	}
}

func TestHandleDoctor(t *testing.T) {
	srv := newTestServer(t)
	tests := []struct {
		method string
		want   int
	}{
		{http.MethodPost, http.StatusOK},
		{http.MethodGet, http.StatusMethodNotAllowed},
		{http.MethodPut, http.StatusMethodNotAllowed},
	}
	for _, tc := range tests {
		t.Run(tc.method, func(t *testing.T) {
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, httptest.NewRequest(tc.method, "/api/v1/doctor", nil))
			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestHandleUpdate(t *testing.T) {
	tests := []struct {
		name   string
		method string
		body   string
		want   int
	}{
		{"happy", http.MethodPost, `{"dry_run":true}`, http.StatusOK},
		{"empty body", http.MethodPost, ``, http.StatusBadRequest},
		{"bad json", http.MethodPost, `{not-json`, http.StatusBadRequest},
		{"wrong method", http.MethodGet, ``, http.StatusMethodNotAllowed},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newTestServer(t)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, "/api/v1/update", strings.NewReader(tc.body))
			srv.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d (body=%s)", rec.Code, tc.want, rec.Body.String())
			}
		})
	}
}

func TestHandleCleanup(t *testing.T) {
	tests := []struct {
		name   string
		method string
		body   string
		want   int
	}{
		{"happy", http.MethodPost, `{"dry_run":true,"all":true}`, http.StatusOK},
		{"bad json", http.MethodPost, `{nope`, http.StatusBadRequest},
		{"wrong method", http.MethodDelete, ``, http.StatusMethodNotAllowed},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newTestServer(t)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, "/api/v1/cleanup", strings.NewReader(tc.body))
			srv.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestHandleSecurityAudit(t *testing.T) {
	srv := newTestServer(t)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/security/audit", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	decodeJSON(t, rec, &resp)
	if resp["success"] != true {
		t.Errorf("success = %v, want true", resp["success"])
	}

	// wrong method
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/security/audit", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET status = %d, want 405", rec.Code)
	}
}

func TestHandleComplianceCheck(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantStd  string
	}{
		{"default", "/api/v1/compliance/check", "cis"},
		{"explicit", "/api/v1/compliance/check?standard=pci", "pci"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newTestServer(t)
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, tc.url, nil))
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			var resp map[string]interface{}
			decodeJSON(t, rec, &resp)
			if resp["standard"] != tc.wantStd {
				t.Errorf("standard = %v, want %s", resp["standard"], tc.wantStd)
			}
		})
	}

	// method not allowed
	srv := newTestServer(t)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/compliance/check", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET status = %d, want 405", rec.Code)
	}
}

func TestHandleHistory(t *testing.T) {
	t.Run("with state", func(t *testing.T) {
		srv := newTestServer(t)
		// seed an operation
		if _, err := srv.state.RecordOperation("test", "seed", "success", "", 0); err != nil {
			t.Fatalf("seed: %v", err)
		}
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/history", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		var resp map[string]interface{}
		decodeJSON(t, rec, &resp)
		ops, ok := resp["operations"].([]interface{})
		if !ok || len(ops) == 0 {
			t.Errorf("expected at least one operation, got %v", resp["operations"])
		}
	})
	t.Run("without state", func(t *testing.T) {
		srv := newTestServerNoState(t)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/history", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})
}

func TestHandlePlugins(t *testing.T) {
	srv := newTestServer(t)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/plugins", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp struct {
		Plugins []map[string]interface{} `json:"plugins"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Plugins) != 5 {
		t.Errorf("plugin count = %d, want 5", len(resp.Plugins))
	}
}

func TestErrorResponse(t *testing.T) {
	srv := newTestServer(t)
	rec := httptest.NewRecorder()
	srv.errorResponse(rec, http.StatusTeapot, "im a teapot")
	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want 418", rec.Code)
	}
	var resp map[string]interface{}
	decodeJSON(t, rec, &resp)
	if resp["message"] != "im a teapot" {
		t.Errorf("message = %v", resp["message"])
	}
	if int(resp["code"].(float64)) != http.StatusTeapot {
		t.Errorf("code = %v", resp["code"])
	}
}

func TestJSONResponse(t *testing.T) {
	srv := newTestServer(t)
	rec := httptest.NewRecorder()
	srv.jsonResponse(rec, http.StatusCreated, map[string]string{"hello": "world"})
	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"hello":"world"`)) {
		t.Errorf("body = %q", rec.Body.String())
	}
}

func TestServerAsHTTPHandler(t *testing.T) {
	// integration: use httptest.Server to verify the Server can act as an http.Handler.
	srv := newTestServer(t)
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}
