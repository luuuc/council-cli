package cmd

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/spf13/cobra"
)

var personasJSON bool

func init() {
	rootCmd.AddCommand(personasCmd)
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
