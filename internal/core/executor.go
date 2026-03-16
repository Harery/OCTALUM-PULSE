// Package core provides core functionality for PULSE
package core

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// Executor handles command execution with optional dry-run and sudo support.
type Executor struct {
	dryRun   bool
	verbose  bool
	sudo     bool
}

// NewExecutor creates a new command executor.
func NewExecutor(dryRun, verbose bool) *Executor {
	return &Executor{
		dryRun:  dryRun,
		verbose: verbose,
		sudo:    os.Geteuid() != 0,
	}
}

// Run executes a command with the given arguments.
func (e *Executor) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := e.buildCommand(ctx, name, args...)

	if e.verbose {
		fmt.Printf("  → Running: %s %s\n", name, strings.Join(args, " "))
	}

	if e.dryRun {
		fmt.Printf("  [DRY-RUN] Would run: %s %s\n", name, strings.Join(args, " "))
		return []byte{}, nil
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("command failed: %s: %w", name, err)
	}

	return output, nil
}

// RunWithSudo executes a command with sudo if not running as root.
func (e *Executor) RunWithSudo(ctx context.Context, name string, args ...string) ([]byte, error) {
	if e.sudo {
		sudoArgs := append([]string{name}, args...)
		return e.Run(ctx, "sudo", sudoArgs...)
	}
	return e.Run(ctx, name, args...)
}

// buildCommand creates an exec.Cmd with context and environment.
func (e *Executor) buildCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	return cmd
}

// Script executes a bash script string.
func (e *Executor) Script(ctx context.Context, script string) ([]byte, error) {
	if e.dryRun {
		fmt.Printf("  [DRY-RUN] Would run script:\n%s\n", script)
		return []byte{}, nil
	}

	cmd := exec.CommandContext(ctx, "bash", "-c", script)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	return cmd.CombinedOutput()
}

// Result represents the outcome of an operation.
type Result struct {
	Success bool
	Output  string
	Error   error
	Changes []Change
}

// Change represents a single modification made during execution.
type Change struct {
	Type        string
	Description string
	Before      string
	After       string
}

// Checker provides system validation utilities.
type Checker struct {
	executor *Executor
}

// NewChecker creates a new system checker.
func NewChecker(executor *Executor) *Checker {
	return &Checker{executor: executor}
}

func (c *Checker) CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func (c *Checker) FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (c *Checker) DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func (c *Checker) IsRoot() bool {
	return os.Geteuid() == 0
}

func (c *Checker) CheckDiskSpace(path string) (uint64, uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0, err
	}
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	return total, free, nil
}
