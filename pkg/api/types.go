// Package api provides API types and structures for OCTALUM-PULSE.
package api

import "time"

// HealthStatus represents the current health state of the system.
type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Platform  string    `json:"platform,omitempty"`
}

// Operation represents a single maintenance operation.
type Operation struct {
	ID          int64     `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Output      string    `json:"output,omitempty"`
	Duration    int64     `json:"duration_ms"`
	Timestamp   time.Time `json:"timestamp"`
}

// OperationResult represents the result of an operation execution.
type OperationResult struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Output  string    `json:"output,omitempty"`
	Changes []Change  `json:"changes,omitempty"`
	Started time.Time `json:"started"`
	Ended   time.Time `json:"ended"`
}

// Change represents a single change made during an operation.
type Change struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Before      string `json:"before,omitempty"`
	After       string `json:"after,omitempty"`
}

// PluginInfo represents metadata about a plugin.
type PluginInfo struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies,omitempty"`
	Enabled      bool     `json:"enabled"`
}

// PlatformInfo represents detected platform information.
type PlatformInfo struct {
	OS             string `json:"os"`
	Distribution   string `json:"distribution"`
	Version        string `json:"version"`
	VersionID      string `json:"version_id"`
	PackageManager string `json:"package_manager"`
	InitSystem     string `json:"init_system"`
	Arch           string `json:"arch"`
	Kernel         string `json:"kernel"`
	Containerized  bool   `json:"containerized"`
	Family         string `json:"family"`
}

// SecurityFinding represents a security issue found during scanning.
type SecurityFinding struct {
	ID          string `json:"id"`
	Severity    string `json:"severity"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Remediation string `json:"remediation,omitempty"`
}

// ComplianceResult represents the result of a compliance check.
type ComplianceResult struct {
	Standard     string           `json:"standard"`
	Compliant    int              `json:"compliant"`
	NonCompliant int              `json:"non_compliant"`
	Skipped      int              `json:"skipped"`
	Findings     []ComplianceItem `json:"findings"`
	Timestamp    time.Time        `json:"timestamp"`
}

// ComplianceItem represents a single compliance check item.
type ComplianceItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Remediation string `json:"remediation,omitempty"`
}

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
