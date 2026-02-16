package cli

import (
	"fmt"
	"os"

	"github.com/jesperpedersen/picky-claude/internal/config"
	"github.com/spf13/cobra"
)

var jsonOutput bool

// rootCmd is the top-level command.
var rootCmd = &cobra.Command{
	Use:   config.BinaryName,
	Short: config.DisplayName + " — quality-enforced development for Claude Code",
	Long: config.DisplayName + ` wraps Claude Code with quality hooks, context
management, spec-driven development, and persistent memory.`,
	Version:       config.Version(),
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
}

// Execute runs the root command. Called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
