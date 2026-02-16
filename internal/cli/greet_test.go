package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestGreetCommand(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"greet"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("greet command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Picky Claude") {
		t.Errorf("output missing product name, got: %q", output)
	}
}

func TestGreetCommandWithName(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"greet", "--name", "Alice"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("greet command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Alice") {
		t.Errorf("output missing name 'Alice', got: %q", output)
	}
}

func TestGreetCommandJSON(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"greet", "--json"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("greet command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"name"`) {
		t.Errorf("JSON output missing 'name' field, got: %q", output)
	}
	if !strings.Contains(output, `"version"`) {
		t.Errorf("JSON output missing 'version' field, got: %q", output)
	}
}
