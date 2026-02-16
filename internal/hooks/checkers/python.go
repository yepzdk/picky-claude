package checkers

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type pythonChecker struct{}

func init() {
	Register(&pythonChecker{})
}

func (c *pythonChecker) Name() string         { return "python" }
func (c *pythonChecker) Extensions() []string { return []string{".py"} }

func (c *pythonChecker) Check(ctx context.Context, filePath string) (*Result, error) {
	result := &Result{}

	if toolExists("ruff") {
		if err := c.runRuff(ctx, filePath, result); err != nil {
			return result, err
		}
	}

	if toolExists("basedpyright") {
		if err := c.runBasedpyright(ctx, filePath, result); err != nil {
			return result, err
		}
	}

	return result, nil
}

func (c *pythonChecker) runRuff(ctx context.Context, filePath string, result *Result) error {
	// ruff check --fix
	cmd := exec.CommandContext(ctx, "ruff", "check", "--fix", "--quiet", filePath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		for _, line := range strings.Split(stderr.String(), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				result.Errors = append(result.Errors, Diagnostic{
					File:    filePath,
					Message: line,
					Source:  "ruff",
				})
			}
		}
	}

	// ruff format
	cmd = exec.CommandContext(ctx, "ruff", "format", "--quiet", filePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ruff format: %w", err)
	}
	result.Fixed = true

	return nil
}

func (c *pythonChecker) runBasedpyright(ctx context.Context, filePath string, result *Result) error {
	cmd := exec.CommandContext(ctx, "basedpyright", filePath)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Run() // basedpyright exits non-zero on errors

	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Found") {
			continue
		}
		if strings.Contains(line, "error:") {
			result.Errors = append(result.Errors, Diagnostic{
				File:    filePath,
				Message: line,
				Source:  "basedpyright",
			})
		} else if strings.Contains(line, "warning:") {
			result.Warnings = append(result.Warnings, Diagnostic{
				File:    filePath,
				Message: line,
				Source:  "basedpyright",
			})
		}
	}
	return nil
}
