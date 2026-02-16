package steps

import (
	"testing"
)

func TestDependencies_Name(t *testing.T) {
	step := &Dependencies{}
	if step.Name() != "dependencies" {
		t.Errorf("expected name 'dependencies', got %q", step.Name())
	}
}

func TestIsInstalled_Git(t *testing.T) {
	// git should be available
	if !isInstalled("git") {
		t.Skip("git not available")
	}
}

func TestIsInstalled_NonExistent(t *testing.T) {
	if isInstalled("nonexistent-binary-xyz-12345") {
		t.Error("expected false for non-existent binary")
	}
}

func TestNpmPackages(t *testing.T) {
	pkgs := npmPackages()
	if len(pkgs) == 0 {
		t.Error("expected at least one npm package")
	}
	// Verify expected packages are in the list
	found := map[string]bool{}
	for _, p := range pkgs {
		found[p.name] = true
	}
	for _, name := range []string{"vexor", "playwright-cli", "mcp-cli"} {
		if !found[name] {
			t.Errorf("expected %s in npm packages", name)
		}
	}
}
