package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(setupRedirectCmd)
}

// setupRedirectCmd is a hidden command that provides a helpful error message
// for users who try to use the deprecated 'council setup' command.
var setupRedirectCmd = &cobra.Command{
	Use:    "setup",
	Short:  "Deprecated: use 'council start' instead",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("'council setup' has been removed\n\nUse 'council start' for zero-config setup")
	},
}
