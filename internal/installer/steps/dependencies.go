package steps

import (
	"fmt"
	"os/exec"

	"github.com/jesperpedersen/picky-claude/internal/installer"
)

// npmPkg describes an npm package to install globally.
type npmPkg struct {
	name    string // Binary name to check in PATH
	pkg     string // npm package name for install
	display string // Human-readable name
}

// npmPackages returns the list of npm packages to install.
func npmPackages() []npmPkg {
	return []npmPkg{
		{name: "vexor", pkg: "@vexor/cli", display: "Vexor (semantic search)"},
		{name: "playwright-cli", pkg: "@anthropic/playwright-cli", display: "Playwright CLI (browser automation)"},
		{name: "mcp-cli", pkg: "@anthropic/mcp-cli", display: "MCP CLI (MCP server access)"},
	}
}

// Dependencies installs required npm packages globally.
type Dependencies struct {
	installed []string // Track what was installed for rollback
}

func (d *Dependencies) Name() string { return "dependencies" }

func (d *Dependencies) Run(ctx *installer.Context) error {
	for _, pkg := range npmPackages() {
		if isInstalled(pkg.name) {
			ctx.Messages = append(ctx.Messages, fmt.Sprintf("  ✓ %s already installed", pkg.display))
			continue
		}

		if err := npmInstallGlobal(pkg.pkg); err != nil {
			return fmt.Errorf("install %s: %w", pkg.display, err)
		}
		d.installed = append(d.installed, pkg.pkg)
		ctx.Messages = append(ctx.Messages, fmt.Sprintf("  + %s installed", pkg.display))
	}
	return nil
}

func (d *Dependencies) Rollback(ctx *installer.Context) {
	for _, pkg := range d.installed {
		npmUninstallGlobal(pkg) //nolint:errcheck
	}
}

// isInstalled checks if a binary exists in PATH.
func isInstalled(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// npmInstallGlobal installs an npm package globally.
func npmInstallGlobal(pkg string) error {
	cmd := exec.Command("npm", "install", "-g", pkg)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("npm install -g %s: %w\n%s", pkg, err, out)
	}
	return nil
}

// npmUninstallGlobal removes a globally installed npm package.
func npmUninstallGlobal(pkg string) error {
	cmd := exec.Command("npm", "uninstall", "-g", pkg)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("npm uninstall -g %s: %w\n%s", pkg, err, out)
	}
	return nil
}
