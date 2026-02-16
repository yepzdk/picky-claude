// Package checkers provides language-specific lint and format tools.
// Each checker detects whether its tools are installed and skips gracefully
// if they are missing.
package checkers

import (
	"context"
	"os/exec"
)

// Diagnostic represents a single error or warning from a checker.
type Diagnostic struct {
	File    string `json:"file"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
	Message string `json:"message"`
	Source  string `json:"source"`
}

// Result holds the output from running a checker on a file.
type Result struct {
	Errors   []Diagnostic `json:"errors,omitempty"`
	Warnings []Diagnostic `json:"warnings,omitempty"`
	Fixed    bool         `json:"fixed"`
}

// Checker is the interface that language-specific checkers implement.
type Checker interface {
	Name() string
	Extensions() []string
	Check(ctx context.Context, filePath string) (*Result, error)
}

// registry holds all registered checkers.
var registry []Checker

// Register adds a checker to the registry.
func Register(c Checker) {
	registry = append(registry, c)
}

// ForExtension returns the first checker that handles the given file extension.
// Returns nil if no checker matches.
func ForExtension(ext string) Checker {
	for _, c := range registry {
		for _, e := range c.Extensions() {
			if e == ext {
				return c
			}
		}
	}
	return nil
}

// toolExists checks if a command-line tool is available on PATH.
func toolExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
