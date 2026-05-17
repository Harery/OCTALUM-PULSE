package api

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestHealthStatusRoundtrip(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	h := HealthStatus{Status: "healthy", Timestamp: now, Version: "1.2.3", Platform: "ubuntu"}
	data, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got HealthStatus
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Status != h.Status || got.Version != h.Version || got.Platform != h.Platform || !got.Timestamp.Equal(h.Timestamp) {
		t.Errorf("mismatch: got=%+v want=%+v", got, h)
	}
}

func TestHealthStatusPlatformOmitempty(t *testing.T) {
	h := HealthStatus{Status: "ok", Version: "v1"}
	data, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if strings.Contains(string(data), "platform") {
		t.Errorf("platform should be omitted: %s", data)
	}
}

func TestOperationRoundtrip(t *testing.T) {
	op := Operation{
		ID: 42, Type: "update", Description: "weekly", Status: "success",
		Output: "ok", Duration: 1500, Timestamp: time.Now().UTC(),
	}
	data, err := json.Marshal(op)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got Operation
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.ID != op.ID || got.Type != op.Type || got.Duration != op.Duration {
		t.Errorf("mismatch: %+v vs %+v", got, op)
	}
	// duration_ms field name check
	if !strings.Contains(string(data), `"duration_ms":1500`) {
		t.Errorf("duration_ms not serialized: %s", data)
	}
}

func TestOperationResultRoundtrip(t *testing.T) {
	r := OperationResult{
		Success: true, Message: "done", Output: "log",
		Changes: []Change{{Type: "pkg", Description: "upgraded", Before: "1.0", After: "1.1"}},
		Started: time.Unix(1, 0).UTC(),
		Ended:   time.Unix(2, 0).UTC(),
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got OperationResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(got.Changes) != 1 || got.Changes[0].Before != "1.0" {
		t.Errorf("changes mismatch: %+v", got.Changes)
	}
}

func TestPluginInfoRoundtrip(t *testing.T) {
	p := PluginInfo{Name: "packages", Version: "2.0", Description: "pkg", Dependencies: []string{"core"}, Enabled: true}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got PluginInfo
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Name != p.Name || got.Enabled != p.Enabled || len(got.Dependencies) != 1 {
		t.Errorf("mismatch: %+v", got)
	}

	// omitempty deps
	p2 := PluginInfo{Name: "x"}
	d2, _ := json.Marshal(p2)
	if strings.Contains(string(d2), "dependencies") {
		t.Errorf("dependencies should be omitted: %s", d2)
	}
}

func TestPlatformInfoRoundtrip(t *testing.T) {
	pi := PlatformInfo{OS: "linux", Distribution: "ubuntu", Version: "22.04", VersionID: "22.04", PackageManager: "apt", InitSystem: "systemd", Arch: "amd64", Kernel: "5.x", Containerized: false, Family: "debian"}
	data, err := json.Marshal(pi)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got PlatformInfo
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got != pi {
		t.Errorf("mismatch: %+v vs %+v", got, pi)
	}
}

func TestSecurityFindingRoundtrip(t *testing.T) {
	f := SecurityFinding{ID: "CVE-1", Severity: "high", Type: "cve", Description: "xx", Remediation: "patch"}
	data, _ := json.Marshal(f)
	var got SecurityFinding
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got != f {
		t.Errorf("mismatch: %+v", got)
	}
}

func TestComplianceResultRoundtrip(t *testing.T) {
	cr := ComplianceResult{
		Standard: "cis", Compliant: 5, NonCompliant: 1, Skipped: 2,
		Findings:  []ComplianceItem{{ID: "1.1", Name: "ssh", Status: "pass", Description: "ok"}},
		Timestamp: time.Now().UTC(),
	}
	data, _ := json.Marshal(cr)
	var got ComplianceResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Standard != cr.Standard || got.Compliant != cr.Compliant || len(got.Findings) != 1 {
		t.Errorf("mismatch: %+v", got)
	}
}

func TestErrorResponseRoundtrip(t *testing.T) {
	e := ErrorResponse{Error: "Bad Request", Message: "invalid input", Code: 400}
	data, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(data), `"code":400`) {
		t.Errorf("code not serialized: %s", data)
	}
	var got ErrorResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got != e {
		t.Errorf("mismatch: %+v vs %+v", got, e)
	}
}
