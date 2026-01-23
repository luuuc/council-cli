package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/luuuc/council-cli/internal/config"
	"github.com/luuuc/council-cli/internal/expert"
	"github.com/spf13/cobra"
)

var listJSON bool
var addYes bool

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)

	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output in JSON format")
	addCmd.Flags().BoolVarP(&addYes, "yes", "y", false, "Skip confirmation prompts")
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all council members",
	Long:  `Shows all experts currently in the council with their ID and focus area.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !config.Exists() {
			return fmt.Errorf("council not initialized: run 'council init' first")
		}

		result, err := expert.ListWithWarnings()
		if err != nil {
			return err
		}

		// JSON output mode
		if listJSON {
			data, err := expert.MarshalExpertsJSON(result.Experts)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		// Display any warnings about files that couldn't be loaded
		for _, warning := range result.Warnings {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", warning)
		}

		if len(result.Experts) == 0 {
			fmt.Println("No experts in the council yet.")
			fmt.Println()
			fmt.Println("Add experts with:")
			fmt.Println("  council setup --apply   (AI-assisted)")
			fmt.Println("  council add \"Name\"      (curated personas)")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tFOCUS")
		for _, e := range result.Experts {
			fmt.Fprintf(w, "%s\t%s\t%s\n", e.ID, e.Name, e.Focus)
		}
		w.Flush()

		return nil
	},
}

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show expert details",
	Long:  `Displays the full details of an expert including their philosophy and principles.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !config.Exists() {
			return fmt.Errorf("council not initialized: run 'council init' first")
		}

		e, err := expert.Load(args[0])
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("expert '%s' not found - run 'council list' to see available experts", args[0])
			}
			return err
		}

		fmt.Printf("ID:    %s\n", e.ID)
		fmt.Printf("Name:  %s\n", e.Name)
		fmt.Printf("Focus: %s\n", e.Focus)

		if len(e.Principles) > 0 {
			fmt.Println("\nPrinciples:")
			for _, p := range e.Principles {
				fmt.Printf("  - %s\n", p)
			}
		}

		if len(e.RedFlags) > 0 {
			fmt.Println("\nRed Flags:")
			for _, r := range e.RedFlags {
				fmt.Printf("  - %s\n", r)
			}
		}

		fmt.Printf("\nFile: %s\n", e.Path())

		return nil
	},
}

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a curated expert to the council",
	Long: `Adds a curated expert from the built-in persona library to your council.

The expert will include pre-written philosophy, principles, and red flags
from the suggestions.yaml database.

For custom experts not in the library, use /council-add with your AI assistant.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !config.Exists() {
			return fmt.Errorf("council not initialized: run 'council init' first")
		}

		name := args[0]

		// Try curated lookup first
		if persona := LookupPersona(name); persona != nil {
			if expert.Exists(persona.ID) {
				return fmt.Errorf("expert '%s' already exists", persona.ID)
			}
			if err := persona.Save(); err != nil {
				return err
			}
			fmt.Printf("Added %s (%s)\n", persona.Name, persona.ID)
			fmt.Printf("File: %s\n", persona.Path())
			fmt.Println()
			fmt.Println("Run 'council sync' to update AI tool configurations.")
			return nil
		}

		// Not found - try suggestion
		if suggestion, distance := SuggestSimilar(name); suggestion != nil {
			// Auto-accept with --yes flag, or prompt for confirmation in interactive mode
			shouldAdd := addYes
			if !shouldAdd && !addYes && isInteractive() && distance <= 2 {
				shouldAdd = Confirm(fmt.Sprintf("Did you mean %q?", suggestion.Name))
			}

			if shouldAdd {
				if expert.Exists(suggestion.ID) {
					return fmt.Errorf("expert '%s' already exists", suggestion.ID)
				}
				if err := suggestion.Save(); err != nil {
					return err
				}
				fmt.Printf("Added %s (%s)\n", suggestion.Name, suggestion.ID)
				fmt.Printf("File: %s\n", suggestion.Path())
				fmt.Println()
				fmt.Println("Run 'council sync' to update AI tool configurations.")
				return nil
			}

			// Non-interactive or user declined: show suggestion in error
			return fmt.Errorf("persona %q not found\n\n"+
				"Did you mean: %s?\n  council add %q\n\n"+
				"Or browse available personas:\n  council personas",
				name, suggestion.Name, suggestion.Name)
		}

		// No close match - helpful error
		return fmt.Errorf("persona %q not found in curated library\n\n"+
			"To create a custom expert, ask your AI:\n  /council-add %s\n\n"+
			"Or browse available personas:\n  council personas", name, name)
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove <id>",
	Short: "Remove an expert from the council",
	Long:  `Removes an expert from the council.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !config.Exists() {
			return fmt.Errorf("council not initialized: run 'council init' first")
		}

		id := args[0]

		e, err := expert.Load(id)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("expert '%s' not found - run 'council list' to see available experts", id)
			}
			return err
		}

		if !Confirm(fmt.Sprintf("Remove %s from the council?", e.Name)) {
			fmt.Println("Cancelled.")
			return nil
		}

		if err := expert.Delete(id); err != nil {
			return err
		}

		fmt.Printf("Removed %s\n", e.Name)

		return nil
	},
}
