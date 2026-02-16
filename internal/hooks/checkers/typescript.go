package checkers

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type typescriptChecker struct{}

func init() {
	Register(&typescriptChecker{})
}

func (c *typescriptChecker) Name() string { return "typescript" }
func (c *typescriptChecker) Extensions() []string {
	return []string{".ts", ".tsx", ".js", ".jsx"}
}

func (c *typescriptChecker) Check(ctx context.Context, filePath string) (*Result, error) {
	result := &Result{}

	if toolExists("prettier") {
		if err := c.runPrettier(ctx, filePath, result); err != nil {
			return result, err
		}
	}

	if toolExists("eslint") {
		if err := c.runEslint(ctx, filePath, result); err != nil {
			return result, err
		}
	}

	if toolExists("tsc") {
		c.runTsc(ctx, filePath, result)
	}

	return result, nil
}

func (c *typescriptChecker) runPrettier(ctx context.Context, filePath string, result *Result) error {
	cmd := exec.CommandContext(ctx, "prettier", "--write", filePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("prettier: %w", err)
	}
	result.Fixed = true
	return nil
}

func (c *typescriptChecker) runEslint(ctx context.Context, filePath string, result *Result) error {
	cmd := exec.CommandContext(ctx, "eslint", "--fix", "--quiet", filePath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		for _, line := range strings.Split(stderr.String(), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				result.Errors = append(result.Errors, Diagnostic{
					File:    filePath,
					Message: line,
					Source:  "eslint",
				})
			}
		}
	}
	return nil
}

func (c *typescriptChecker) runTsc(ctx context.Context, filePath string, result *Result) {
	cmd := exec.CommandContext(ctx, "tsc", "--noEmit", "--pretty", "false")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Run()

	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, "error TS") {
			result.Errors = append(result.Errors, Diagnostic{
				File:    filePath,
				Message: line,
				Source:  "tsc",
			})
		}
	}
}
