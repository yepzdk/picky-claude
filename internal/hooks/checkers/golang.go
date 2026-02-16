package checkers

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

type golangChecker struct{}

func init() {
	Register(&golangChecker{})
}

func (c *golangChecker) Name() string         { return "go" }
func (c *golangChecker) Extensions() []string { return []string{".go"} }

func (c *golangChecker) Check(ctx context.Context, filePath string) (*Result, error) {
	result := &Result{}

	if toolExists("gofmt") {
		if err := c.runGofmt(ctx, filePath, result); err != nil {
			return result, err
		}
	}

	if toolExists("golangci-lint") {
		c.runGolangciLint(ctx, filePath, result)
	}

	return result, nil
}

func (c *golangChecker) runGofmt(ctx context.Context, filePath string, result *Result) error {
	cmd := exec.CommandContext(ctx, "gofmt", "-w", filePath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		result.Errors = append(result.Errors, Diagnostic{
			File:    filePath,
			Message: stderr.String(),
			Source:  "gofmt",
		})
		return nil
	}
	result.Fixed = true
	return nil
}

func (c *golangChecker) runGolangciLint(ctx context.Context, filePath string, result *Result) {
	cmd := exec.CommandContext(ctx, "golangci-lint", "run", "--new-from-rev=HEAD", filePath)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Run()

	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		result.Warnings = append(result.Warnings, Diagnostic{
			File:    filePath,
			Message: line,
			Source:  "golangci-lint",
		})
	}
}
