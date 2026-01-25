package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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
			fmt.Println("  council add \"Name\"    Add from curated library or create custom")
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
	Short: "Add expert to council (from library or create custom)",
	Long: `Adds an expert to your council.

If the name matches a curated expert from the library, adds it directly.
If no match is found, guides you through creating a custom expert.

Examples:
  council add "Kent Beck"     # Found in library - adds directly
  council add "My CTO"        # Not found - creates custom persona
  council add "rob pike"      # Case-insensitive matching`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !config.Exists() {
			return fmt.Errorf("council not initialized: run 'council start' first")
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
			if !shouldAdd && isInteractive() && distance <= 2 {
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
		}

		// No match found - trigger creation flow
		if !isInteractive() {
			return fmt.Errorf("persona %q not found in curated library\n\n"+
				"To create a custom expert interactively, run without piping:\n  council add %q\n\n"+
				"Or browse available personas:\n  council personas", name, name)
		}

		fmt.Printf("'%s' not found in curated library. Let's create a custom persona.\n\n", name)
		return runAddCreationFlow(name)
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

// runAddCreationFlow guides the user through creating a custom expert
// for the project council (.council/experts/).
func runAddCreationFlow(name string) error {
	reader := bufio.NewReader(os.Stdin)

	// Generate ID from name
	id := expert.ToID(name)

	// Check if expert already exists
	if expert.Exists(id) {
		return fmt.Errorf("expert '%s' already exists", id)
	}

	// Focus (required)
	fmt.Print("Focus (one-line description of their expertise): ")
	focus, _ := reader.ReadString('\n')
	focus = trimNewline(focus)
	if focus == "" {
		return fmt.Errorf("focus is required")
	}

	// Philosophy (optional)
	fmt.Print("Philosophy (optional, press Enter to skip): ")
	philosophy, _ := reader.ReadString('\n')
	philosophy = trimNewline(philosophy)

	// Create expert
	e := &expert.Expert{
		ID:         id,
		Name:       name,
		Focus:      focus,
		Philosophy: philosophy,
	}

	// Save to project council
	if err := e.Save(); err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("Created %s (%s)\n", e.Name, e.ID)
	fmt.Printf("File: %s\n", e.Path())
	fmt.Println()
	fmt.Println("Run 'council sync' to update AI tool configurations.")

	return nil
}

// trimNewline removes trailing newline characters from a string
func trimNewline(s string) string {
	return strings.TrimRight(s, "\r\n")
}
