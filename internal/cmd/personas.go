package cmd

import (
	"cmp"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/luuuc/council-cli/internal/creator"
	"github.com/luuuc/council-cli/internal/expert"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

//go:embed suggestions.yaml
var suggestionsYAML []byte

// suggestionBank holds all expert suggestions loaded from YAML
var suggestionBank map[string][]expert.Expert

func init() {
	if err := yaml.Unmarshal(suggestionsYAML, &suggestionBank); err != nil {
		panic(fmt.Sprintf("failed to parse suggestions.yaml: %v", err))
	}
}

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

		// List experts in installed repo
		experts, _ := creator.ListInstalledExperts()
		var repoExperts []*expert.Expert
		for _, e := range experts {
			if e.Source == "installed:"+name {
				repoExperts = append(repoExperts, e)
			}
		}

		fmt.Println()
		fmt.Printf("Installed to: %s\n\n", name)

		if len(repoExperts) > 0 {
			fmt.Println("Experts added:")
			for _, e := range repoExperts {
				fmt.Printf("  - %s - %s\n", e.ID, e.Focus)
			}
			fmt.Println()
		}

		fmt.Println("Personas available via:")
		fmt.Println("  council personas --json")

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

		// Get experts for each installed repo
		experts, _ := creator.ListInstalledExperts()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "REPOSITORY\tEXPERTS")

		for _, name := range installed {
			count := 0
			for _, e := range experts {
				if e.Source == "installed:"+name {
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

// LookupPersona finds a curated persona by name or ID (case-insensitive).
// Returns nil if not found.
func LookupPersona(nameOrID string) *expert.Expert {
	normalized := strings.ToLower(strings.TrimSpace(nameOrID))

	for _, experts := range suggestionBank {
		for _, e := range experts {
			// Match by ID
			if strings.ToLower(e.ID) == normalized {
				copy := e
				return &copy
			}
			// Match by name (case-insensitive)
			if strings.ToLower(e.Name) == normalized {
				copy := e
				return &copy
			}
			// Match by name converted to ID format (spaces → dashes)
			if strings.ToLower(strings.ReplaceAll(e.Name, " ", "-")) == normalized {
				copy := e
				return &copy
			}
		}
	}
	return nil
}

// levenshtein computes the edit distance between two strings.
func levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	d := make([][]int, len(a)+1)
	for i := range d {
		d[i] = make([]int, len(b)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			d[i][j] = min(d[i-1][j]+1, d[i][j-1]+1, d[i-1][j-1]+cost)
		}
	}
	return d[len(a)][len(b)]
}

// SuggestSimilar finds the closest persona match using edit distance.
// Returns nil if no close match (distance > 3), if exact match exists,
// or if the input is too short to match reliably.
// The second return value is the edit distance of the match.
func SuggestSimilar(input string) (*expert.Expert, int) {
	// If LookupPersona would find this, don't suggest
	if LookupPersona(input) != nil {
		return nil, 0
	}

	normalized := strings.ToLower(strings.TrimSpace(input))

	// Require minimum input length to avoid false matches
	if len(normalized) < 4 {
		return nil, 0
	}

	var bestMatch *expert.Expert
	bestDistance := 4 // Threshold: only consider distance <= 3

	for _, experts := range suggestionBank {
		for _, e := range experts {
			// Check distance against name
			if d := levenshtein(normalized, strings.ToLower(e.Name)); d < bestDistance && d > 0 {
				bestDistance = d
				copy := e
				bestMatch = &copy
			}
			// Check distance against ID
			if d := levenshtein(normalized, strings.ToLower(e.ID)); d < bestDistance && d > 0 {
				bestDistance = d
				copy := e
				bestMatch = &copy
			}
		}
	}
	return bestMatch, bestDistance
}
