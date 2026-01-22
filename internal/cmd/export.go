package cmd

import (
	"fmt"

	"github.com/luuuc/council-cli/internal/config"
	"github.com/luuuc/council-cli/internal/export"
	"github.com/luuuc/council-cli/internal/expert"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export council as portable markdown",
	Long: `Exports your council as clean markdown for use anywhere.

The output can be:
- Pasted into any AI chat
- Used as custom instructions in desktop apps
- Saved to a file for sharing
- Piped to clipboard with pbcopy/xclip

Examples:
  council export              # Output to stdout
  council export | pbcopy     # Copy to clipboard (macOS)
  council export | xclip      # Copy to clipboard (Linux)
  council export > council.md # Save to file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !config.Exists() {
			return fmt.Errorf("council not initialized: run 'council init' first")
		}

		experts, err := expert.List()
		if err != nil {
			return err
		}

		if len(experts) == 0 {
			return fmt.Errorf("no experts to export - add some with 'council add' or 'council setup --apply'")
		}

		fmt.Print(export.FormatMarkdown(experts))
		return nil
	},
}
