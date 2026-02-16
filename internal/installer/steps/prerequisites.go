package steps

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/jesperpedersen/picky-claude/internal/installer"
)

// Prerequisites checks that required tools are installed and the OS is supported.
type Prerequisites struct{}

func (p *Prerequisites) Name() string { return "prerequisites" }

func (p *Prerequisites) Run(ctx *installer.Context) error {
	if err := checkOS(); err != nil {
		return err
	}

	required := []struct {
		cmd  string
		args string
		name string
	}{
		{"git", "--version", "git"},
		{"claude", "--version", "Claude Code"},
	}

	for _, r := range required {
		if err := checkCommand(r.cmd, r.args); err != nil {
			return fmt.Errorf("%s is required but not found in PATH: %w", r.name, err)
		}
	}

	return nil
}

func (p *Prerequisites) Rollback(ctx *installer.Context) {
	// Nothing to undo — this step only checks, doesn't modify.
}

// checkCommand verifies a command exists and runs without error.
func checkCommand(name, arg string) error {
	cmd := exec.Command(name, arg)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", name, arg, err)
	}
	return nil
}

// checkOS verifies the operating system is supported.
func checkOS() error {
	switch runtime.GOOS {
	case "darwin", "linux":
		return nil
	default:
		return fmt.Errorf("unsupported OS: %s (supported: darwin, linux)", runtime.GOOS)
	}
}
