package main

import (
	"os"
	"testing"

	"OCTALUM-PULSE/internal/platform"
)

func init() {
	if platform.DetectedPlatform == nil {
		platform.DetectedPlatform = &platform.Info{
			OS: "linux", Distribution: "ubuntu", Version: "22.04", VersionID: "22.04",
			PackageManager: "apt", InitSystem: "systemd", Arch: "amd64", Kernel: "test",
		}
	}
}

// withArgs temporarily replaces os.Args for the duration of fn.
func withArgs(t *testing.T, args []string, fn func()) {
	t.Helper()
	orig := os.Args
	t.Cleanup(func() { os.Args = orig })
	os.Args = args
	fn()
}

// withEnv temporarily sets env vars and restores them after the test.
func withEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for k, v := range kv {
		old, hadOld := os.LookupEnv(k)
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("setenv: %v", err)
		}
		t.Cleanup(func() {
			if hadOld {
				_ = os.Setenv(k, old)
			} else {
				_ = os.Unsetenv(k)
			}
		})
	}
}

func TestRun_Help(t *testing.T) {
	// Isolate config paths to a temp dir so config.Load doesn't read /etc.
	tmp := t.TempDir()
	withEnv(t, map[string]string{
		"HOME":            tmp,
		"XDG_CONFIG_HOME": tmp,
		"XDG_DATA_HOME":   tmp,
	})
	withArgs(t, []string{"pulse", "--help"}, func() {
		got := run()
		if got != 0 {
			t.Errorf("run() = %d, want 0 for --help", got)
		}
	})
}

func TestRun_VersionFlag(t *testing.T) {
	tmp := t.TempDir()
	withEnv(t, map[string]string{
		"HOME":            tmp,
		"XDG_CONFIG_HOME": tmp,
		"XDG_DATA_HOME":   tmp,
	})
	withArgs(t, []string{"pulse", "--version"}, func() {
		// cobra by default prints version and exits zero when --version is wired in
		_ = run()
	})
}

func TestRun_UnknownFlag(t *testing.T) {
	tmp := t.TempDir()
	withEnv(t, map[string]string{
		"HOME":            tmp,
		"XDG_CONFIG_HOME": tmp,
		"XDG_DATA_HOME":   tmp,
	})
	withArgs(t, []string{"pulse", "--definitely-not-a-flag"}, func() {
		got := run()
		if got == 0 {
			t.Errorf("run() = 0, want non-zero for unknown flag")
		}
	})
}
