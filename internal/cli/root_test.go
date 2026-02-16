package cli

import (
	"bytes"
	"testing"
)

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestRootHelp(t *testing.T) {
	out, err := executeCommand("--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) == 0 {
		t.Error("expected help output, got empty string")
	}
}

func TestVersion(t *testing.T) {
	out, err := executeCommand("--version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) == 0 {
		t.Error("expected version output, got empty string")
	}
}

func TestSubcommandsExist(t *testing.T) {
	commands := []string{
		"run", "serve", "install", "hook", "session",
		"worktree", "check-context", "send-clear",
		"register-plan", "greet", "statusline",
	}
	for _, name := range commands {
		t.Run(name, func(t *testing.T) {
			_, err := executeCommand(name, "--help")
			if err != nil {
				t.Errorf("command %q returned error: %v", name, err)
			}
		})
	}
}

func TestWorktreeSubcommandsExist(t *testing.T) {
	subs := []string{"create", "detect", "diff", "sync", "cleanup", "status"}
	for _, name := range subs {
		t.Run(name, func(t *testing.T) {
			_, err := executeCommand("worktree", name, "--help")
			if err != nil {
				t.Errorf("worktree %q returned error: %v", name, err)
			}
		})
	}
}
