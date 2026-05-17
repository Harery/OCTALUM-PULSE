package models

import (
	"encoding/json"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestConfigJSONRoundtrip(t *testing.T) {
	original := Config{
		Version:  1,
		LogLevel: "debug",
		Plugins: Plugins{
			Packages:      PluginConfig{Enabled: true, Options: []string{"security-only"}},
			Security:      PluginConfig{Enabled: true},
			Performance:   PluginConfig{Enabled: false},
			Compliance:    PluginConfig{Enabled: false, Options: []string{"cis", "pci"}},
			Observability: PluginConfig{Enabled: true, Options: []string{"prom"}},
		},
		AI:    AIConfig{Enabled: true, Mode: "remote", Model: "claude"},
		Cloud: CloudConfig{Enabled: true, Endpoint: "https://example.com"},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got Config
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.Version != original.Version || got.LogLevel != original.LogLevel {
		t.Errorf("scalar mismatch: got %+v", got)
	}
	if got.Plugins.Packages.Enabled != true || len(got.Plugins.Packages.Options) != 1 {
		t.Errorf("plugins not preserved: %+v", got.Plugins.Packages)
	}
	if got.AI.Mode != "remote" || got.Cloud.Endpoint != "https://example.com" {
		t.Errorf("nested config not preserved: AI=%+v Cloud=%+v", got.AI, got.Cloud)
	}
}

func TestConfigYAMLRoundtrip(t *testing.T) {
	original := Config{
		Version:  2,
		LogLevel: "warn",
		Plugins:  Plugins{Packages: PluginConfig{Enabled: true}},
		AI:       AIConfig{Mode: "local"},
		Cloud:    CloudConfig{Enabled: false},
	}

	data, err := yaml.Marshal(&original)
	if err != nil {
		t.Fatalf("yaml.Marshal: %v", err)
	}

	var got Config
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatalf("yaml.Unmarshal: %v", err)
	}
	if got.Version != 2 || got.LogLevel != "warn" || got.AI.Mode != "local" {
		t.Errorf("yaml roundtrip mismatch: %+v", got)
	}
}

func TestConfigOmitEmpty(t *testing.T) {
	// Options has omitempty in json; empty slice should drop on json output.
	pc := PluginConfig{Enabled: true}
	data, err := json.Marshal(pc)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := string(data); got != `{"enabled":true}` {
		t.Errorf("omitempty failed: got %s", got)
	}
}

func TestMetricRoundtrip(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	m := Metric{Name: "cpu", Value: 42.5, Unit: "percent", Type: "gauge", Timestamp: now}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got Metric
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Name != m.Name || got.Value != m.Value || got.Unit != m.Unit ||
		got.Type != m.Type || !got.Timestamp.Equal(m.Timestamp) {
		t.Errorf("roundtrip mismatch: got=%+v want=%+v", got, m)
	}
}

func TestSnapshotRoundtrip(t *testing.T) {
	s := Snapshot{ID: 7, Name: "pre-upgrade", Description: "before X", Data: "{}", Timestamp: time.Now().UTC()}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got Snapshot
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.ID != s.ID || got.Name != s.Name || got.Description != s.Description || got.Data != s.Data {
		t.Errorf("snapshot mismatch: got=%+v want=%+v", got, s)
	}
}

func TestZeroValueRoundtrip(t *testing.T) {
	// Empty config should marshal and unmarshal cleanly.
	var zero Config
	data, err := json.Marshal(zero)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got Config
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Version != 0 || got.LogLevel != "" || got.AI.Enabled {
		t.Errorf("zero-value mismatch: %+v", got)
	}
}
