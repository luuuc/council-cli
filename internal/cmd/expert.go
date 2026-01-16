package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/luuuc/council-cli/internal/config"
	"github.com/luuuc/council-cli/internal/expert"
	"github.com/spf13/cobra"
)

var addFocus string

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)

	addCmd.Flags().StringVar(&addFocus, "focus", "", "Expert's focus area (required)")
	addCmd.MarkFlagRequired("focus")
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all council members",
	Long:  `Shows all experts currently in the council with their ID and focus area.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !config.Exists() {
			return fmt.Errorf("council not initialized: run 'council init' first")
		}

		experts, err := expert.List()
		if err != nil {
			return err
		}

		if len(experts) == 0 {
			fmt.Println("No experts in the council yet.")
			fmt.Println()
			fmt.Println("Add experts with:")
			fmt.Println("  council setup --apply   (AI-assisted)")
			fmt.Println("  council add \"Name\" --focus \"area\"")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tFOCUS")
		for _, e := range experts {
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
				return fmt.Errorf("expert '%s' not found", args[0])
			}
			return err
		}

		fmt.Printf("ID:    %s\n", e.ID)
		fmt.Printf("Name:  %s\n", e.Name)
		fmt.Printf("Focus: %s\n", e.Focus)

		if len(e.Triggers.Paths) > 0 {
			fmt.Println("\nPaths:")
			for _, p := range e.Triggers.Paths {
				fmt.Printf("  %s\n", p)
			}
		}

		if len(e.Triggers.Keywords) > 0 {
			fmt.Println("\nKeywords:")
			for _, k := range e.Triggers.Keywords {
				fmt.Printf("  %s\n", k)
			}
		}

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
	Short: "Add an expert to the council",
	Long:  `Creates a new expert with the given name and focus area.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !config.Exists() {
			return fmt.Errorf("council not initialized: run 'council init' first")
		}

		name := args[0]
		id := expert.ToID(name)

		if expert.Exists(id) {
			return fmt.Errorf("expert '%s' already exists", id)
		}

		e := &expert.Expert{
			ID:    id,
			Name:  name,
			Focus: addFocus,
		}

		if err := e.Save(); err != nil {
			return err
		}

		fmt.Printf("Added %s (%s)\n", e.Name, e.ID)
		fmt.Printf("Edit: %s\n", e.Path())

		return nil
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
				return fmt.Errorf("expert '%s' not found", id)
			}
			return err
		}

		if !expert.Confirm(fmt.Sprintf("Remove %s from the council?", e.Name)) {
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
