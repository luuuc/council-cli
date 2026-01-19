package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"text/tabwriter"

	"github.com/luuuc/council-cli/internal/creator"
	"github.com/spf13/cobra"
)

var (
	creatorNewInterview bool
	creatorNewFrom      string
)

func init() {
	rootCmd.AddCommand(creatorCmd)
	creatorCmd.AddCommand(creatorInitCmd)
	creatorCmd.AddCommand(creatorListCmd)
	creatorCmd.AddCommand(creatorShowCmd)
	creatorCmd.AddCommand(creatorDeleteCmd)
	creatorCmd.AddCommand(creatorNewCmd)
	creatorCmd.AddCommand(creatorEditCmd)
	creatorCmd.AddCommand(creatorRemoteCmd)
	creatorCmd.AddCommand(creatorPushCmd)
	creatorCmd.AddCommand(creatorPullCmd)
	creatorCmd.AddCommand(creatorPublishCmd)

	creatorNewCmd.Flags().BoolVar(&creatorNewInterview, "interview", false, "AI-assisted persona creation")
	creatorNewCmd.Flags().StringVar(&creatorNewFrom, "from", "", "Fork from existing persona ID")
}

var creatorCmd = &cobra.Command{
	Use:   "creator",
	Short: "Manage your personal council of custom personas (advanced)",
	Long: `Create and manage custom expert personas in your personal council.

NOTE: This is an advanced feature for power users who want to create
and share custom expert personas. Most users should use the standard
setup commands instead.

Your personal council is stored in a git-tracked directory:
  macOS:   ~/Library/Application Support/council/my-council/
  Linux:   ~/.config/council/my-council/
  Windows: %AppData%\council\my-council\

Quick start:
  council creator init       Initialize personal council
  council creator new        Create a new persona (interactive)
  council creator list       List all custom personas
  council creator show <id>  Show persona details`,
}

var creatorInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize your personal council",
	Long:  `Creates a git-tracked directory for your custom personas.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := creator.Init(); err != nil {
			return err
		}

		path, err := creator.MyCouncilPath()
		if err != nil {
			return fmt.Errorf("failed to get my-council path: %w", err)
		}
		fmt.Printf("Initialized personal council at:\n  %s\n\n", path)
		fmt.Println("Next steps:")
		fmt.Println("  council creator new           Create your first persona")
		fmt.Println("  council creator new --interview   AI-assisted creation")
		return nil
	},
}

var creatorListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all custom personas",
	Long:  `Shows all personas in your personal council.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !creator.Initialized() {
			return fmt.Errorf("personal council not initialized: run 'council creator init' first")
		}

		experts, err := creator.List()
		if err != nil {
			return err
		}

		if len(experts) == 0 {
			fmt.Println("No custom personas yet.")
			fmt.Println()
			fmt.Println("Create one with:")
			fmt.Println("  council creator new")
			fmt.Println("  council creator new --interview   (AI-assisted)")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tFOCUS\tPRIORITY")
		for _, e := range experts {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.ID, e.Name, e.Focus, e.Priority)
		}
		w.Flush()

		return nil
	},
}

var creatorShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show persona details",
	Long:  `Displays the full details of a custom persona.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !creator.Initialized() {
			return fmt.Errorf("personal council not initialized: run 'council creator init' first")
		}

		e, err := creator.Load(args[0])
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("persona '%s' not found - run 'council creator list' to see your personas", args[0])
			}
			return err
		}

		fmt.Printf("ID:       %s\n", e.ID)
		fmt.Printf("Name:     %s\n", e.Name)
		fmt.Printf("Focus:    %s\n", e.Focus)
		fmt.Printf("Category: %s\n", e.Category)
		fmt.Printf("Priority: %s\n", e.Priority)

		if len(e.Triggers) > 0 {
			fmt.Println("\nTriggers:")
			for _, t := range e.Triggers {
				fmt.Printf("  - %s\n", t)
			}
		}

		if e.Philosophy != "" {
			fmt.Printf("\nPhilosophy:\n  %s\n", e.Philosophy)
		}

		if len(e.Principles) > 0 {
			fmt.Println("\nPrinciples:")
			for _, pr := range e.Principles {
				fmt.Printf("  - %s\n", pr)
			}
		}

		if len(e.RedFlags) > 0 {
			fmt.Println("\nRed Flags:")
			for _, rf := range e.RedFlags {
				fmt.Printf("  - %s\n", rf)
			}
		}

		if path, err := creator.ExpertPath(e.ID); err == nil {
			fmt.Printf("\nFile: %s\n", path)
		}

		return nil
	},
}

var creatorDeleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Aliases: []string{"remove", "rm"},
	Short:   "Delete a custom persona",
	Long:    `Removes a persona from your personal council and commits the change.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !creator.Initialized() {
			return fmt.Errorf("personal council not initialized: run 'council creator init' first")
		}

		id := args[0]

		// Load to get name for confirmation
		e, err := creator.Load(id)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("persona '%s' not found - run 'council creator list' to see your personas", id)
			}
			return err
		}

		// Confirm
		if !Confirm(fmt.Sprintf("Delete '%s'?", e.Name)) {
			fmt.Println("Cancelled.")
			return nil
		}

		if err := creator.Delete(id); err != nil {
			return err
		}

		fmt.Printf("Deleted %s\n", e.Name)
		return nil
	},
}

var creatorNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new custom persona",
	Long: `Create a new expert persona in your personal council.

Modes:
  (default)     Interactive wizard to define persona
  --interview   AI-assisted persona creation from description
  --from <id>   Fork from an existing persona`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !creator.Initialized() {
			return fmt.Errorf("personal council not initialized: run 'council creator init' first")
		}

		// Fork mode
		if creatorNewFrom != "" {
			return runForkMode(creatorNewFrom)
		}

		// Interview mode
		if creatorNewInterview {
			return runInterviewMode()
		}

		// Default: interactive wizard
		return runInteractiveWizard()
	},
}

var creatorEditCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit a custom expert",
	Long:  `Opens the expert file in your $EDITOR. Changes are auto-committed.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !creator.Initialized() {
			return fmt.Errorf("personal council not initialized: run 'council creator init' first")
		}

		id := args[0]

		// Check exists
		e, err := creator.Load(id)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("expert '%s' not found - run 'council creator list' to see your personas", id)
			}
			return err
		}

		path, err := creator.ExpertPath(e.ID)
		if err != nil {
			return fmt.Errorf("failed to get expert path: %w", err)
		}

		// Open in editor
		if err := openInEditor(path); err != nil {
			return err
		}

		// Reload to verify it's still valid
		e, err = creator.Load(id)
		if err != nil {
			return fmt.Errorf("file has errors after editing: %w", err)
		}

		// Commit changes
		myCouncil, err := creator.MyCouncilPath()
		if err != nil {
			return fmt.Errorf("failed to get my-council path: %w", err)
		}
		repo := creator.NewGitRepo(myCouncil)
		if err := repo.AddAndCommit(id+".md", fmt.Sprintf("Update expert: %s", e.Name)); err != nil {
			return err
		}

		fmt.Printf("Updated %s\n", e.Name)
		return nil
	},
}

// openInEditor opens a file in the user's configured editor.
func openInEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		if _, err := exec.LookPath("vim"); err == nil {
			editor = "vim"
		} else if _, err := exec.LookPath("nano"); err == nil {
			editor = "nano"
		} else if _, err := exec.LookPath("notepad"); err == nil {
			editor = "notepad"
		} else {
			return fmt.Errorf("no editor found: set $EDITOR environment variable")
		}
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
