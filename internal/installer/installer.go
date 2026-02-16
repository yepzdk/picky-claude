// Package installer implements a multi-step project installer with rollback.
// Each step performs a discrete setup action and can undo its changes if a
// later step fails.
package installer

import (
	"fmt"
	"io"
	"path/filepath"
)

// Step is a single installation action that can be run and rolled back.
type Step interface {
	Name() string
	Run(ctx *Context) error
	Rollback(ctx *Context)
}

// Context carries shared state across installation steps.
type Context struct {
	ProjectDir string // Root of the project being set up
	ClaudeDir  string // .claude/ directory path
	HomeDir    string // User's home directory
	Messages   []string
}

// Result holds the outcome of an installation run.
type Result struct {
	Success    bool
	Completed  []string // Names of steps that completed successfully
	FailedStep string   // Name of the step that failed (empty on success)
	Error      error    // The error from the failed step
}

// Installer runs a sequence of Steps with rollback on failure.
type Installer struct {
	projectDir string
	steps      []Step
}

// New creates an Installer for the given project directory with the given steps.
func New(projectDir string, steps ...Step) *Installer {
	return &Installer{
		projectDir: projectDir,
		steps:      steps,
	}
}

// Run executes each step in order. If a step fails, all previously completed
// steps are rolled back in reverse order.
func (inst *Installer) Run() *Result {
	ctx := inst.context()
	var completed []Step

	for _, step := range inst.steps {
		if err := step.Run(ctx); err != nil {
			// Rollback completed steps in reverse order
			for i := len(completed) - 1; i >= 0; i-- {
				completed[i].Rollback(ctx)
			}

			names := make([]string, len(completed))
			for i, s := range completed {
				names[i] = s.Name()
			}

			return &Result{
				Success:    false,
				Completed:  names,
				FailedStep: step.Name(),
				Error:      fmt.Errorf("step %q failed: %w", step.Name(), err),
			}
		}
		completed = append(completed, step)
	}

	names := make([]string, len(completed))
	for i, s := range completed {
		names[i] = s.Name()
	}

	return &Result{
		Success:   true,
		Completed: names,
	}
}

// RunWithUI executes the installer with terminal UI output.
func (inst *Installer) RunWithUI(w io.Writer) *Result {
	ui := NewUI(w)
	ui.Banner()

	ctx := inst.context()
	total := len(inst.steps)
	var completed []Step

	for i, step := range inst.steps {
		ui.StepStart(i+1, total, step.Name())
		ctx.Messages = nil // Reset per-step messages

		if err := step.Run(ctx); err != nil {
			ui.StepFailed(step.Name(), err)
			ui.RollingBack()

			for j := len(completed) - 1; j >= 0; j-- {
				completed[j].Rollback(ctx)
			}

			names := make([]string, len(completed))
			for j, s := range completed {
				names[j] = s.Name()
			}

			result := &Result{
				Success:    false,
				Completed:  names,
				FailedStep: step.Name(),
				Error:      fmt.Errorf("step %q failed: %w", step.Name(), err),
			}
			ui.Summary(result)
			return result
		}

		ui.Messages(ctx.Messages)
		ui.StepDone(step.Name())
		completed = append(completed, step)
	}

	names := make([]string, len(completed))
	for i, s := range completed {
		names[i] = s.Name()
	}

	result := &Result{
		Success:   true,
		Completed: names,
	}
	ui.Summary(result)
	return result
}

// context builds the shared Context from the installer's configuration.
func (inst *Installer) context() *Context {
	return &Context{
		ProjectDir: inst.projectDir,
		ClaudeDir:  filepath.Join(inst.projectDir, ".claude"),
	}
}
