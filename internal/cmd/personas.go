package cmd

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"text/tabwriter"

	"github.com/luuuc/council-cli/internal/creator"
	"github.com/spf13/cobra"
)

var personasJSON bool

func init() {
	rootCmd.AddCommand(personasCmd)
	personasCmd.AddCommand(personasInstallCmd)
	personasCmd.AddCommand(personasInstalledCmd)
	personasCmd.AddCommand(personasUpdateCmd)
	personasCmd.AddCommand(personasUninstallCmd)

	personasCmd.Flags().BoolVar(&personasJSON, "json", false, "Output as JSON")
}

// PersonaJSON is a flattened persona with category
type PersonaJSON struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Category   string   `json:"category"`
	Focus      string   `json:"focus"`
	Philosophy string   `json:"philosophy,omitempty"`
	Principles []string `json:"principles,omitempty"`
	RedFlags   []string `json:"red_flags,omitempty"`
	Triggers   []string `json:"triggers,omitempty"`
}

var personasCmd = &cobra.Command{
	Use:   "personas",
	Short: "List all available expert personas",
	Long:  `Shows all built-in expert personas that can be added to a council.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var personas []PersonaJSON

		for category, experts := range suggestionBank {
			for _, e := range experts {
				personas = append(personas, PersonaJSON{
					ID:         e.ID,
					Name:       e.Name,
					Category:   category,
					Focus:      e.Focus,
					Philosophy: e.Philosophy,
					Principles: e.Principles,
					RedFlags:   e.RedFlags,
					Triggers:   e.Triggers,
				})
			}
		}

		// Sort by category, then name for deterministic output
		slices.SortFunc(personas, func(a, b PersonaJSON) int {
			if c := cmp.Compare(a.Category, b.Category); c != 0 {
				return c
			}
			return cmp.Compare(a.Name, b.Name)
		})

		if personasJSON {
			data, err := json.MarshalIndent(personas, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		// Human-readable output
		fmt.Printf("Available personas (%d):\n\n", len(personas))
		for _, p := range personas {
			fmt.Printf("  %s (%s) - %s\n", p.Name, p.Category, p.Focus)
		}

		return nil
	},
}

var personasInstallCmd = &cobra.Command{
	Use:   "install <url>",
	Short: "Install personas from a git repository",
	Long: `Clones a git repository containing custom personas.

The repository will be cloned to your local config directory and
its personas will be available in council setup and sync.

Examples:
  council personas install https://github.com/dhh/my-council.git
  council personas install git@github.com:dhh/my-council.git`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		fmt.Printf("Installing from %s...\n\n", url)

		name, err := creator.Install(url)
		if err != nil {
			return err
		}

		// List personas in installed repo
		personas, _ := creator.ListInstalledPersonas()
		var repoPersonas []*creator.Persona
		for _, p := range personas {
			if p.Source == "installed:"+name {
				repoPersonas = append(repoPersonas, p)
			}
		}

		fmt.Println()
		fmt.Printf("Installed to: %s\n\n", name)

		if len(repoPersonas) > 0 {
			fmt.Println("Personas added:")
			for _, p := range repoPersonas {
				fmt.Printf("  - %s - %s\n", p.ID, p.Focus)
			}
			fmt.Println()
		}

		fmt.Println("Use in setup:")
		fmt.Println("  council setup --interactive")

		return nil
	},
}

var personasInstalledCmd = &cobra.Command{
	Use:   "installed",
	Short: "List installed persona repositories",
	Long:  `Shows all persona repositories installed from remote sources.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		installed, err := creator.ListInstalled()
		if err != nil {
			return err
		}

		if len(installed) == 0 {
			fmt.Println("No repositories installed.")
			fmt.Println()
			fmt.Println("Install one with:")
			fmt.Println("  council personas install <url>")
			return nil
		}

		// Get personas for each installed repo
		personas, _ := creator.ListInstalledPersonas()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "REPOSITORY\tPERSONAS")

		for _, name := range installed {
			count := 0
			for _, p := range personas {
				if p.Source == "installed:"+name {
					count++
				}
			}
			fmt.Fprintf(w, "%s\t%d\n", name, count)
		}

		w.Flush()
		return nil
	},
}

var personasUpdateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update installed persona repositories",
	Long: `Pulls the latest changes from installed repositories.

If no name is specified, updates all installed repositories.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			// Update specific repo
			name := args[0]
			fmt.Printf("Updating %s...\n", name)

			if err := creator.Update(name); err != nil {
				return err
			}

			fmt.Println("Done")
			return nil
		}

		// Update all
		fmt.Println("Updating all installed repositories...")
		fmt.Println()

		updated, err := creator.UpdateAll()
		if err != nil {
			return err
		}

		if len(updated) == 0 {
			fmt.Println("No repositories to update.")
			return nil
		}

		for _, name := range updated {
			fmt.Printf("  Updated %s\n", name)
		}

		fmt.Println()
		fmt.Println("Done")
		return nil
	},
}

var personasUninstallCmd = &cobra.Command{
	Use:   "uninstall <name>",
	Short: "Uninstall a persona repository",
	Long:  `Removes an installed persona repository.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if !Confirm(fmt.Sprintf("Uninstall '%s'?", name)) {
			fmt.Println("Cancelled.")
			return nil
		}

		if err := creator.Uninstall(name); err != nil {
			return err
		}

		fmt.Printf("Uninstalled %s\n", name)
		return nil
	},
}
