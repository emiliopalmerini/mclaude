package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "claude-watcher",
	Short: "Analytics and experimentation platform for Claude Code",
	Long: `Claude Watcher is a personal analytics and experimentation platform for Claude Code usage.

Track sessions, measure token consumption, estimate costs, and run experiments
to optimize your Claude Code workflow.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(recordCmd)
	rootCmd.AddCommand(migrateCmd)
}
