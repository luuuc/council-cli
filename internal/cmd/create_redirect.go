package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createRedirectCmd)
}

// createRedirectCmd is a hidden command that provides a helpful error message
// for users who try to use the deprecated 'council create' command.
// The create functionality was merged into 'council add' in Cycle 1.
var createRedirectCmd = &cobra.Command{
	Use:    "create",
	Short:  "Deprecated: use 'council add' instead",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("'council create' has been merged into 'council add'\n\nUse 'council add \"Name\"' to add or create a persona")
	},
}
