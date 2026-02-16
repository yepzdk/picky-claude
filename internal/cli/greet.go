package cli

import (
	"encoding/json"
	"fmt"

	"github.com/jesperpedersen/picky-claude/internal/config"
	"github.com/spf13/cobra"
)

var greetName string

const banner = `
 ____  _      _
|  _ \(_) ___| | ___   _
| |_) | |/ __| |/ / | | |
|  __/| | (__|   <| |_| |
|_|   |_|\___|_|\_\\__, |
                    |___/
`

var greetCmd = &cobra.Command{
	Use:   "greet",
	Short: "Print the welcome banner",
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if jsonOutput {
			data := map[string]string{
				"name":    config.DisplayName,
				"version": config.Version(),
			}
			if greetName != "" {
				data["user"] = greetName
			}
			return json.NewEncoder(out).Encode(data)
		}

		fmt.Fprint(out, banner)
		fmt.Fprintf(out, "  %s v%s\n", config.DisplayName, config.Version())

		if greetName != "" {
			fmt.Fprintf(out, "  Welcome, %s!\n", greetName)
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "  Quality-enforced development for Claude Code")
		fmt.Fprintln(out)

		return nil
	},
}

func init() {
	greetCmd.Flags().StringVar(&greetName, "name", "", "name to greet")
	rootCmd.AddCommand(greetCmd)
}
