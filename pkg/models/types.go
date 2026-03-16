// Package models provides data models and configuration structures for OCTALUM-PULSE.
package models

import "time"

// Config represents the main configuration file structure.
type Config struct {
	Version  int         `yaml:"version" json:"version"`
	LogLevel string      `yaml:"log_level" json:"log_level"`
	Plugins  Plugins     `yaml:"plugins" json:"plugins"`
	AI       AIConfig    `yaml:"ai" json:"ai"`
	Cloud    CloudConfig `yaml:"cloud" json:"cloud"`
}

// Plugins contains configuration for all plugins.
type Plugins struct {
	Packages      PluginConfig `yaml:"packages" json:"packages"`
	Security      PluginConfig `yaml:"security" json:"security"`
	Performance   PluginConfig `yaml:"performance" json:"performance"`
	Compliance    PluginConfig `yaml:"compliance" json:"compliance"`
	Observability PluginConfig `yaml:"observability" json:"observability"`
}

// PluginConfig represents configuration for a single plugin.
type PluginConfig struct {
	Enabled bool     `yaml:"enabled" json:"enabled"`
	Options []string `yaml:"options,omitempty" json:"options,omitempty"`
}

// AIConfig represents AI integration configuration.
type AIConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Mode    string `yaml:"mode" json:"mode"`
	Model   string `yaml:"model" json:"model"`
}

// CloudConfig represents cloud integration configuration.
type CloudConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	Endpoint string `yaml:"endpoint" json:"endpoint"`
}

// Metric represents a single metric data point.
type Metric struct {
	Name      string    `json:"name"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

// Snapshot represents a state snapshot for backup/restore.
type Snapshot struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Data        string    `json:"data"`
	Timestamp   time.Time `json:"timestamp"`
}
