package cli

import (
	"os"

	"github.com/jesperpedersen/picky-claude/internal/installer"
	"github.com/jesperpedersen/picky-claude/internal/installer/steps"
	"github.com/spf13/cobra"
)

var skipPrereqs bool
var skipDeps bool

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Set up project with rules, hooks, and configuration",
	Long: `Runs a multi-step installer that sets up the project with Claude Code
rules, hooks, configuration files, and recommended extensions.

Each step supports rollback on failure.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		var installSteps []installer.Step

		if !skipPrereqs {
			installSteps = append(installSteps, &steps.Prerequisites{})
		}
		if !skipDeps {
			installSteps = append(installSteps, &steps.Dependencies{})
		}

		installSteps = append(installSteps,
			&steps.ShellConfig{},
			&steps.ClaudeFiles{},
			&steps.ConfigFiles{},
			&steps.VSCode{},
			&steps.Finalize{},
		)

		inst := installer.New(dir, installSteps...)
		result := inst.RunWithUI(cmd.OutOrStdout())

		if !result.Success {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	installCmd.Flags().BoolVar(&skipPrereqs, "skip-prereqs", false, "skip prerequisite checks")
	installCmd.Flags().BoolVar(&skipDeps, "skip-deps", false, "skip dependency installation")
	rootCmd.AddCommand(installCmd)
}
