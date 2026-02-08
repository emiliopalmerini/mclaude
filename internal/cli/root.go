package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var app *AppContext

// skipAppContext lists commands that manage raw DB schema and don't need repositories.
var skipAppContext = map[string]bool{
	"migrate": true,
	"reset":   true,
	"help":    true,
	"mclaude": true,
	"record":  true, // record manages its own DB for test overrides
	"hook":    true, // hook manages its own DB like record
}

var rootCmd = &cobra.Command{
	Use:   "mclaude",
	Short: "Analytics and experimentation platform for Claude Code",
	Long: `mclaude is a personal analytics and experimentation platform for Claude Code usage.

Track sessions, measure token consumption, estimate costs, and run experiments
to optimize your Claude Code workflow.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if skipAppContext[cmd.Name()] {
			return nil
		}
		var err error
		app, err = NewAppContext()
		return err
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if app != nil {
			return app.Close()
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(recordCmd)
	rootCmd.AddCommand(hookCmd)
	rootCmd.AddCommand(migrateCmd)
}
