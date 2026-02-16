package assets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListAssets_Rules(t *testing.T) {
	files, err := ListAssets("rules")
	if err != nil {
		t.Fatalf("ListAssets(rules) failed: %v", err)
	}
	if len(files) == 0 {
		t.Error("expected at least one rule file")
	}
	// Should not include .gitkeep
	for _, f := range files {
		if filepath.Base(f) == ".gitkeep" {
			t.Errorf("ListAssets should not include .gitkeep, got: %s", f)
		}
	}
}

func TestListAssets_Commands(t *testing.T) {
	files, err := ListAssets("commands")
	if err != nil {
		t.Fatalf("ListAssets(commands) failed: %v", err)
	}
	if len(files) == 0 {
		t.Error("expected at least one command file")
	}
}

func TestListAssets_Agents(t *testing.T) {
	files, err := ListAssets("agents")
	if err != nil {
		t.Fatalf("ListAssets(agents) failed: %v", err)
	}
	if len(files) == 0 {
		t.Error("expected at least one agent file")
	}
}

func TestListAssets_InvalidCategory(t *testing.T) {
	_, err := ListAssets("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent category")
	}
}

func TestReadAsset(t *testing.T) {
	data, err := ReadAsset("rules/example.md")
	if err != nil {
		t.Fatalf("ReadAsset failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty content")
	}
}

func TestExtractTo(t *testing.T) {
	dir := t.TempDir()
	targetDir := filepath.Join(dir, ".claude")

	err := ExtractTo(targetDir)
	if err != nil {
		t.Fatalf("ExtractTo failed: %v", err)
	}

	// Check rules were extracted
	ruleFile := filepath.Join(targetDir, "rules", "example.md")
	if _, err := os.Stat(ruleFile); os.IsNotExist(err) {
		t.Errorf("expected %s to exist after extraction", ruleFile)
	}

	// Check commands
	cmdFile := filepath.Join(targetDir, "commands", "spec.md")
	if _, err := os.Stat(cmdFile); os.IsNotExist(err) {
		t.Errorf("expected %s to exist after extraction", cmdFile)
	}

	// Check agents
	agentFile := filepath.Join(targetDir, "agents", "plan-verifier.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Errorf("expected %s to exist after extraction", agentFile)
	}

	// Verify content matches
	data, _ := os.ReadFile(ruleFile)
	embedded, _ := ReadAsset("rules/example.md")
	if string(data) != string(embedded) {
		t.Error("extracted content does not match embedded content")
	}
}

func TestExtractTo_Idempotent(t *testing.T) {
	dir := t.TempDir()
	targetDir := filepath.Join(dir, ".claude")

	// Extract twice — should not error
	if err := ExtractTo(targetDir); err != nil {
		t.Fatalf("first ExtractTo failed: %v", err)
	}
	if err := ExtractTo(targetDir); err != nil {
		t.Fatalf("second ExtractTo failed: %v", err)
	}
}
