package cmd

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/luuuc/council-cli/internal/detect"
	"github.com/luuuc/council-cli/internal/expert"
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

// ExpertSuggestion represents a suggested expert with selection state
type ExpertSuggestion struct {
	Expert   expert.Expert
	Selected bool
}

// InteractiveSetup runs the interactive council setup flow
func InteractiveSetup() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Council Setup")
	fmt.Println("=============")
	fmt.Println()

	// Phase 1: Intention
	intention := askIntention(reader)
	fmt.Println()

	// Phase 2: Detect project
	fmt.Println("Scanning project...")
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}
	detection, err := detect.Scan(dir)
	if err != nil {
		return fmt.Errorf("detection failed: %w", err)
	}

	if detection.Summary() != "" {
		fmt.Printf("Detected: %s\n", detection.Summary())
	} else {
		fmt.Println("No specific stack detected.")
	}
	fmt.Println()

	// Phase 3: Suggest experts based on detection and intention
	suggestions := suggestExperts(detection, intention)

	if len(suggestions) == 0 {
		fmt.Println("No automatic suggestions available for this stack.")
		fmt.Println("Use 'council add <name>' to add experts manually.")
		return nil
	}

	// Phase 4: Let user select experts
	selected := selectExperts(reader, suggestions)

	if len(selected) == 0 {
		fmt.Println("No experts selected. Setup cancelled.")
		return nil
	}

	// Phase 5: Confirm and create
	fmt.Println()
	fmt.Printf("Creating council with %d experts:\n", len(selected))
	for _, e := range selected {
		fmt.Printf("  - %s\n", e.Name)
	}
	fmt.Println()

	if !Confirm("Proceed?") {
		fmt.Println("Cancelled.")
		return nil
	}

	// Create expert files
	for _, e := range selected {
		if err := e.Save(); err != nil {
			return fmt.Errorf("failed to save %s: %w", e.Name, err)
		}
		fmt.Printf("Created %s\n", e.Path())
	}

	fmt.Println()
	fmt.Println("Council created! Next steps:")
	fmt.Println("  council list    View your council")
	fmt.Println("  council sync    Sync to AI tool configs")

	return nil
}

func askIntention(reader *bufio.Reader) string {
	fmt.Println("What do you want your council for?")
	fmt.Println("  1. Code review")
	fmt.Println("  2. Writing (docs, blog, book)")
	fmt.Println("  3. Business decisions")
	fmt.Println("  4. A bit of everything")
	fmt.Println()
	fmt.Print("Choice [1]: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "2":
		return "writing"
	case "3":
		return "business"
	case "4":
		return "everything"
	default:
		return "code"
	}
}

func selectExperts(reader *bufio.Reader, suggestions []ExpertSuggestion) []expert.Expert {
	// Initially all are selected
	for i := range suggestions {
		suggestions[i].Selected = true
	}

	for {
		fmt.Println("Suggested experts (toggle with number, done with Enter):")
		fmt.Println()

		for i, s := range suggestions {
			mark := "[x]"
			if !s.Selected {
				mark = "[ ]"
			}
			fmt.Printf("  %d. %s %s - %s\n", i+1, mark, s.Expert.Name, s.Expert.Focus)
		}

		fmt.Println()
		fmt.Print("Toggle (1-" + strconv.Itoa(len(suggestions)) + ") or Enter to continue: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			break
		}

		num, err := strconv.Atoi(input)
		if err != nil || num < 1 || num > len(suggestions) {
			fmt.Println("Invalid selection.")
			continue
		}

		suggestions[num-1].Selected = !suggestions[num-1].Selected
		fmt.Println()
	}

	var selected []expert.Expert
	for _, s := range suggestions {
		if s.Selected {
			selected = append(selected, s.Expert)
		}
	}
	return selected
}

// suggestExperts returns expert suggestions based on detected stack and intention
func suggestExperts(d *detect.Detection, intention string) []ExpertSuggestion {
	var suggestions []ExpertSuggestion
	seen := make(map[string]bool)

	add := func(experts []expert.Expert) {
		for _, e := range experts {
			if !seen[e.ID] {
				seen[e.ID] = true
				suggestions = append(suggestions, ExpertSuggestion{Expert: e})
			}
		}
	}

	// Check for frameworks first
	hasRails := false
	hasPhoenix := false
	for _, fw := range d.Frameworks {
		switch fw.Name {
		case "Rails":
			hasRails = true
			add(suggestionBank["rails"])
		case "Phoenix":
			hasPhoenix = true
			add(suggestionBank["phoenix"])
		}
	}

	// Language-specific experts
	for _, lang := range d.Languages {
		switch lang.Name {
		case "Go":
			add(suggestionBank["go"])
		case "Ruby":
			if !hasRails {
				add(suggestionBank["ruby"])
			}
		case "Elixir":
			if !hasPhoenix {
				add(suggestionBank["elixir"])
			}
		case "Python":
			add(suggestionBank["python"])
		case "TypeScript", "JavaScript":
			add(suggestionBank["javascript"])
			add(suggestionBank["frontend"])
		case "Rust":
			add(suggestionBank["rust"])
		}
	}

	// Add general experts for code projects
	if intention == "code" || intention == "everything" {
		add(suggestionBank["general"])
	}

	// Writing experts
	if intention == "writing" {
		add(suggestionBank["writing"])
	}

	// Business experts
	if intention == "business" {
		add(suggestionBank["business"])
		add(suggestionBank["general"]) // Jason Fried fits business too
	}

	return suggestions
}
