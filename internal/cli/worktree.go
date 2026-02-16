package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jesperpedersen/picky-claude/internal/worktree"
	"github.com/spf13/cobra"
)

// repoDir returns the current working directory as the git repository root.
func repoDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	return dir, nil
}

var worktreeCmd = &cobra.Command{
	Use:   "worktree",
	Short: "Git worktree management for isolated development",
}

var worktreeCreateCmd = &cobra.Command{
	Use:   "create <slug>",
	Short: "Create an isolated git worktree",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := repoDir()
		if err != nil {
			return err
		}

		mgr := worktree.NewManager(dir)
		info, err := mgr.Create(args[0])
		if err != nil {
			return err
		}

		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(info)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Created worktree at %s (branch: %s, base: %s)\n",
			info.Path, info.Branch, info.BaseBranch)
		return nil
	},
}

var worktreeDetectCmd = &cobra.Command{
	Use:   "detect <slug>",
	Short: "Check if a worktree exists",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := repoDir()
		if err != nil {
			return err
		}

		mgr := worktree.NewManager(dir)
		info, err := mgr.Detect(args[0])
		if err != nil {
			return err
		}

		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(info)
		}
		if info.Found {
			fmt.Fprintf(cmd.OutOrStdout(), "Worktree found at %s (branch: %s)\n", info.Path, info.Branch)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "Worktree not found")
		}
		return nil
	},
}

var worktreeDiffCmd = &cobra.Command{
	Use:   "diff <slug>",
	Short: "List changed files in a worktree",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := repoDir()
		if err != nil {
			return err
		}

		mgr := worktree.NewManager(dir)
		result, err := mgr.Diff(args[0])
		if err != nil {
			return err
		}

		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
		}
		if len(result.Files) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No changed files")
			return nil
		}
		for _, f := range result.Files {
			fmt.Fprintln(cmd.OutOrStdout(), f)
		}
		return nil
	},
}

var worktreeSyncCmd = &cobra.Command{
	Use:   "sync <slug>",
	Short: "Squash merge worktree changes to base branch",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := repoDir()
		if err != nil {
			return err
		}

		mgr := worktree.NewManager(dir)
		result, err := mgr.Sync(args[0])
		if err != nil {
			return err
		}

		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Synced %d files (commit: %s)\n",
			result.FilesChanged, result.CommitHash)
		return nil
	},
}

var worktreeCleanupCmd = &cobra.Command{
	Use:   "cleanup <slug>",
	Short: "Remove a worktree and its branch",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := repoDir()
		if err != nil {
			return err
		}

		mgr := worktree.NewManager(dir)
		if err := mgr.Cleanup(args[0]); err != nil {
			return err
		}

		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]bool{"success": true})
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Worktree cleaned up")
		return nil
	},
}

var worktreeStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show active worktree info",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := repoDir()
		if err != nil {
			return err
		}

		mgr := worktree.NewManager(dir)
		status, err := mgr.Status()
		if err != nil {
			return err
		}

		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(status)
		}
		if !status.Active {
			fmt.Fprintln(cmd.OutOrStdout(), "No active worktree")
			return nil
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Active worktree: %s (branch: %s, base: %s)\n",
			status.Slug, status.Branch, status.BaseBranch)
		return nil
	},
}

func init() {
	worktreeCmd.AddCommand(
		worktreeCreateCmd,
		worktreeDetectCmd,
		worktreeDiffCmd,
		worktreeSyncCmd,
		worktreeCleanupCmd,
		worktreeStatusCmd,
	)
	rootCmd.AddCommand(worktreeCmd)
}
