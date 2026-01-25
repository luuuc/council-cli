package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/luuuc/council-cli/internal/adapter"
	"github.com/luuuc/council-cli/internal/config"
	"github.com/luuuc/council-cli/internal/detect"
	"github.com/luuuc/council-cli/internal/expert"
	"github.com/luuuc/council-cli/internal/sync"
	"github.com/spf13/cobra"
)

const (
	maxStackExperts = 3 // Maximum stack-specific experts to add
	maxTotalExperts = 5 // Maximum total experts in auto-selection
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "First-time setup (zero-config)",
	Long: `Sets up your council with sensible defaults. One command, zero decisions.

What it does:
  1. Creates .council/ directory
  2. Detects your project's stack
  3. Adds 5 experts based on your stack
  4. Syncs to your AI tool

If you already have a council, use 'council add' to add more experts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStart()
	},
}

func runStart() error {
	// Check if .council/ already exists
	if config.Exists() {
		return fmt.Errorf(".council/ already exists\n\nTo modify your council:\n  council add <name>    Add an expert\n  council remove <id>   Remove an expert\n  council list          View your council")
	}

	// Step 1: Detect AI tool
	tool, err := detectTool()
	if err != nil {
		return err
	}

	// Step 2: Create directory structure (same as init)
	dirs := []string{
		config.CouncilDir,
		config.Path(config.ExpertsDir),
		config.Path(config.CommandsDir),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create %s: %w", dir, err)
		}
	}

	// Create config with detected tool
	cfg := config.Default()
	cfg.Tool = tool
	if err := cfg.Save(); err != nil {
		return err
	}

	// Create .gitkeep files
	for _, subdir := range []string{config.ExpertsDir, config.CommandsDir} {
		path := config.Path(subdir, ".gitkeep")
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create .gitkeep: %w", err)
		}
	}

	// Get adapter for display name
	a, _ := adapter.Get(tool)
	displayName := tool
	if a != nil {
		displayName = a.DisplayName()
	}
	fmt.Printf("✓ Detected: %s\n", displayName)

	// Step 3: Detect project stack
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	d, err := detect.Scan(dir)
	if err != nil {
		return fmt.Errorf("detection failed: %w", err)
	}

	fmt.Printf("✓ Detected: %s\n", d.Summary())

	// Step 4: Select experts based on detected stack
	experts := selectExperts(d)
	if len(experts) == 0 {
		// Fallback to generalists if detection returned nothing useful
		experts = selectGeneralists()
	}

	// Step 5: Add selected experts
	var added []*expert.Expert
	for _, e := range experts {
		if err := e.Save(); err != nil {
			fmt.Printf("  Warning: failed to add %s: %v\n", e.Name, err)
			continue
		}
		added = append(added, e)
	}

	if len(added) == 0 {
		return fmt.Errorf("failed to add any experts")
	}

	// Print added experts
	var names []string
	for _, e := range added {
		names = append(names, e.Name)
	}
	fmt.Printf("✓ Added %d experts: %s\n", len(added), joinNames(names))

	// Step 6: Sync to AI tool
	if err := sync.SyncAll(cfg, sync.Options{}); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	fmt.Println()
	fmt.Println("Your council is ready. Try: /council <topic>")

	return nil
}

// detectTool determines which AI tool to use (auto-detect, single tool, or first of multiple)
func detectTool() (string, error) {
	detected := adapter.Detect()

	switch len(detected) {
	case 0:
		return "generic", nil
	case 1:
		return detected[0].Name(), nil
	default:
		// Multiple tools - use first one (deterministic order)
		return detected[0].Name(), nil
	}
}

// selectExperts picks up to 5 experts based on detected stack
func selectExperts(d *detect.Detection) []*expert.Expert {
	var selected []*expert.Expert
	seen := make(map[string]bool)

	// Map categories from detection to suggestion bank categories
	categories := mapDetectionToCategories(d)

	// Add stack-specific experts
	for _, cat := range categories {
		if len(selected) >= maxStackExperts {
			break
		}
		if experts, ok := loadSuggestionBank()[cat]; ok && len(experts) > 0 {
			e := &experts[0] // Get first (primary) expert from category
			if !seen[e.ID] {
				selected = append(selected, expertFromSuggestion(e))
				seen[e.ID] = true
			}
		}
	}

	// Always try to add generalists to round out the council
	generalists := []string{"kent-beck", "jason-fried", "dieter-rams"}
	for _, id := range generalists {
		if len(selected) >= maxTotalExperts {
			break
		}
		if seen[id] {
			continue
		}
		if e := findExpertByID(id); e != nil {
			selected = append(selected, e)
			seen[id] = true
		}
	}

	return selected
}

// selectGeneralists returns default generalists when detection finds nothing
func selectGeneralists() []*expert.Expert {
	var selected []*expert.Expert
	ids := []string{"kent-beck", "dieter-rams", "jason-fried", "sandi-metz", "cal-newport"}

	for _, id := range ids {
		if len(selected) >= maxTotalExperts {
			break
		}
		if e := findExpertByID(id); e != nil {
			selected = append(selected, e)
		}
	}

	return selected
}

// mapDetectionToCategories maps detected stack to suggestion bank categories
func mapDetectionToCategories(d *detect.Detection) []string {
	var categories []string

	// Map languages
	for _, lang := range d.Languages {
		switch lang.Name {
		case "Go":
			categories = append(categories, "go")
		case "Ruby":
			categories = append(categories, "ruby")
		case "Python":
			categories = append(categories, "python")
		case "JavaScript", "TypeScript":
			categories = append(categories, "javascript")
		case "Rust":
			categories = append(categories, "rust")
		case "Elixir":
			categories = append(categories, "elixir")
		case "Java", "Kotlin":
			categories = append(categories, "java")
		case "C#":
			categories = append(categories, "dotnet")
		case "Swift":
			categories = append(categories, "swift")
		}
	}

	// Map frameworks
	for _, fw := range d.Frameworks {
		switch fw.Name {
		case "Rails":
			categories = append(categories, "rails")
		case "Phoenix":
			categories = append(categories, "elixir")
		case "Django", "Flask", "FastAPI":
			categories = append(categories, "python")
		case "React", "Vue", "Next.js":
			categories = append(categories, "frontend")
		case "Express":
			categories = append(categories, "javascript")
		}
	}

	// Map testing to add testing expert
	if len(d.Testing) > 0 {
		categories = append(categories, "testing")
	}

	return categories
}

// findExpertByID searches the suggestion bank for an expert by ID
func findExpertByID(id string) *expert.Expert {
	for _, experts := range loadSuggestionBank() {
		for i := range experts {
			if experts[i].ID == id {
				return expertFromSuggestion(&experts[i])
			}
		}
	}
	return nil
}

// expertFromSuggestion converts a suggestion bank expert to an expert.Expert
func expertFromSuggestion(e *expert.Expert) *expert.Expert {
	return &expert.Expert{
		ID:         e.ID,
		Name:       e.Name,
		Focus:      e.Focus,
		Philosophy: e.Philosophy,
		Principles: e.Principles,
		RedFlags:   e.RedFlags,
		Triggers:   e.Triggers,
	}
}

// joinNames joins names with commas
func joinNames(names []string) string {
	return strings.Join(names, ", ")
}
