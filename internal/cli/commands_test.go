//go:build linux

package cli

import (
	"bytes"
	"context"
	"testing"

	"OCTALUM-PULSE/internal/config"
	"OCTALUM-PULSE/internal/version"
)

// runCommand executes the root cobra command with the given args and returns
// (combined stdout/stderr buffer, exec error).
func runCommand(t *testing.T, args []string) (string, error) {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.Database.Path = t.TempDir() + "/test.db"
	root := NewRootCommand(cfg, version.Info{Version: "test", GitCommit: "abc", BuildTime: "now"})

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestSubcommandsHelp(t *testing.T) {
	subs := []string{
		"doctor", "fix", "update", "cleanup", "security", "compliance",
		"history", "rollback", "explain", "plugin", "tui", "version",
	}
	for _, sub := range subs {
		t.Run(sub, func(t *testing.T) {
			out, err := runCommand(t, []string{sub, "--help"})
			if err != nil {
				t.Errorf("%s --help: %v\nout=%s", sub, err, out)
			}
			if out == "" {
				t.Errorf("%s --help produced no output", sub)
			}
		})
	}
}

func TestVersionCommand(t *testing.T) {
	// version cmd uses fmt.Println which writes directly to os.Stdout, so
	// we cannot capture the output via cobra's SetOut. Just verify it
	// executes without error.
	if _, err := runCommand(t, []string{"version"}); err != nil {
		t.Fatalf("version cmd: %v", err)
	}
}

func TestRootHelp(t *testing.T) {
	out, err := runCommand(t, []string{"--help"})
	if err != nil {
		t.Fatalf("--help: %v", err)
	}
	if out == "" {
		t.Error("--help produced no output")
	}
}

func TestUnknownCommand(t *testing.T) {
	_, err := runCommand(t, []string{"nope-not-a-command"})
	if err == nil {
		t.Error("unknown command should return error")
	}
}

func TestApp_OutputJSON(t *testing.T) {
	cfg := config.DefaultConfig()
	app, err := NewApp(cfg)
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}
	defer app.Close()
	// outputJSON writes to stdout — call it for coverage; we can't capture
	// without redirecting os.Stdout, but we can ensure it doesn't error.
	if err := app.outputJSON(map[string]string{"a": "b"}); err != nil {
		t.Errorf("outputJSON: %v", err)
	}
}

func TestApp_RegisterBuiltInPlugins(t *testing.T) {
	cfg := config.DefaultConfig()
	app, err := NewApp(cfg)
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}
	defer app.Close()
	app.registerBuiltInPlugins() // no-op but exercises the function
}

func TestApp_RunDoctorChecks(t *testing.T) {
	cfg := config.DefaultConfig()
	app, err := NewApp(cfg)
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}
	defer app.Close()

	results := app.runDoctorChecks(context.Background(), true)
	if results == nil {
		t.Error("runDoctorChecks returned nil")
	}
}

func TestApp_AttemptFix(t *testing.T) {
	cfg := config.DefaultConfig()
	app, err := NewApp(cfg)
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}
	defer app.Close()

	// Just verify it returns a bool without panicking on a benign issue.
	_ = app.attemptFix(context.Background(), HealthResult{Name: "test", Status: "warn"})
}
