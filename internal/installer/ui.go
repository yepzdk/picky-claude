package installer

import (
	"fmt"
	"io"

	"github.com/jesperpedersen/picky-claude/internal/config"
)

// UI handles terminal output for the installation process.
type UI struct {
	w io.Writer
}

// NewUI creates a UI that writes to the given writer.
func NewUI(w io.Writer) *UI {
	return &UI{w: w}
}

// Banner prints the installation header.
func (u *UI) Banner() {
	fmt.Fprintf(u.w, "\n%s Installer\n", config.DisplayName)
	fmt.Fprintln(u.w, "──────────────────────────────")
}

// StepStart prints a message when a step begins.
func (u *UI) StepStart(num, total int, name string) {
	fmt.Fprintf(u.w, "\n[%d/%d] %s...\n", num, total, name)
}

// StepDone prints a success message after a step completes.
func (u *UI) StepDone(name string) {
	fmt.Fprintf(u.w, "  ✓ %s complete\n", name)
}

// StepFailed prints an error message when a step fails.
func (u *UI) StepFailed(name string, err error) {
	fmt.Fprintf(u.w, "  ✗ %s failed: %v\n", name, err)
}

// Messages prints accumulated context messages from a step.
func (u *UI) Messages(msgs []string) {
	for _, msg := range msgs {
		fmt.Fprintln(u.w, msg)
	}
}

// RollingBack prints a rollback notification.
func (u *UI) RollingBack() {
	fmt.Fprintln(u.w, "\nRolling back completed steps...")
}

// Summary prints the final installation result.
func (u *UI) Summary(result *Result) {
	fmt.Fprintln(u.w)
	if result.Success {
		fmt.Fprintf(u.w, "✓ %s installed successfully!\n", config.DisplayName)
		fmt.Fprintf(u.w, "  Run `%s run` to start a session.\n", config.BinaryName)
	} else {
		fmt.Fprintf(u.w, "✗ Installation failed at step %q\n", result.FailedStep)
		fmt.Fprintf(u.w, "  Error: %v\n", result.Error)
		if len(result.Completed) > 0 {
			fmt.Fprintf(u.w, "  Rolled back: %v\n", result.Completed)
		}
	}
	fmt.Fprintln(u.w)
}
