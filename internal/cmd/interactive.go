package cmd

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/luuuc/council-cli/internal/creator"
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
	Source   string // "custom", "installed:<name>", or "" for built-in
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

	// Phase 3: Load custom experts (always first)
	var customSuggestions []ExpertSuggestion
	customExperts, _ := creator.List()
	for _, e := range customExperts {
		customSuggestions = append(customSuggestions, ExpertSuggestion{
			Expert:   *e,
			Selected: e.Priority == "always",
			Source:   "custom",
		})
	}

	// Phase 3b: Load installed experts
	var installedSuggestions []ExpertSuggestion
	installedExperts, _ := creator.ListInstalledExperts()
	for _, e := range installedExperts {
		installedSuggestions = append(installedSuggestions, ExpertSuggestion{
			Expert:   *e,
			Selected: e.Priority == "always",
			Source:   e.Source,
		})
	}

	// Phase 4: Suggest built-in experts based on detection and intention
	suggestions := suggestExperts(detection, intention)

	// Combine all suggestions
	allSuggestions := append(customSuggestions, installedSuggestions...)
	allSuggestions = append(allSuggestions, suggestions...)

	if len(allSuggestions) == 0 {
		fmt.Println("No automatic suggestions available for this stack.")
		fmt.Println("Use 'council add <name>' to add experts manually.")
		return nil
	}

	// Phase 5: Let user select experts
	selected := selectExpertsWithSources(reader, allSuggestions, len(customSuggestions), len(installedSuggestions))

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

// Selection limits
const (
	maxPerCategory = 2
	maxTotal       = 7
	preSelectCount = 4
)

// suggestExperts returns expert suggestions based on detected stack and intention
func suggestExperts(d *detect.Detection, intention string) []ExpertSuggestion {
	var suggestions []ExpertSuggestion
	seen := make(map[string]bool)

	// addN adds up to n experts from a category
	addN := func(experts []expert.Expert, n int) int {
		added := 0
		for _, e := range experts {
			if added >= n {
				break
			}
			if !seen[e.ID] {
				seen[e.ID] = true
				suggestions = append(suggestions, ExpertSuggestion{Expert: e})
				added++
			}
		}
		return added
	}

	// Build pattern set for trigger matching
	patterns := buildPatternSet(d)

	// Phase 1: Framework experts (highest priority, up to 2)
	frameworkCount := 0
	for _, fw := range d.Frameworks {
		if frameworkCount >= maxPerCategory {
			break
		}
		switch fw.Name {
		case "Rails":
			frameworkCount += addN(suggestionBank["rails"], maxPerCategory-frameworkCount)
		case "Phoenix":
			frameworkCount += addN(suggestionBank["phoenix"], maxPerCategory-frameworkCount)
		case "Next.js", "React", "Vue":
			frameworkCount += addN(suggestionBank["frontend"], maxPerCategory-frameworkCount)
		}
	}

	// Phase 2: Language experts (up to 2, skip if framework covers it)
	langCount := 0
	coveredByFramework := map[string]bool{
		"Ruby":   hasFramework(d, "Rails"),
		"Elixir": hasFramework(d, "Phoenix"),
	}

	for _, lang := range d.Languages {
		if langCount >= maxPerCategory {
			break
		}
		if coveredByFramework[lang.Name] {
			continue
		}
		switch lang.Name {
		case "Go":
			langCount += addN(suggestionBank["go"], maxPerCategory-langCount)
		case "Ruby":
			langCount += addN(suggestionBank["ruby"], maxPerCategory-langCount)
		case "Elixir":
			langCount += addN(suggestionBank["elixir"], maxPerCategory-langCount)
		case "Python":
			langCount += addN(suggestionBank["python"], maxPerCategory-langCount)
		case "TypeScript", "JavaScript":
			langCount += addN(suggestionBank["javascript"], maxPerCategory-langCount)
		case "Rust":
			langCount += addN(suggestionBank["rust"], maxPerCategory-langCount)
		case "Swift":
			langCount += addN(suggestionBank["swift"], maxPerCategory-langCount)
		case "Kotlin":
			langCount += addN(suggestionBank["kotlin"], maxPerCategory-langCount)
		}
	}

	// Phase 3: General experts based on intention
	if intention == "code" || intention == "everything" {
		// Add core general experts first
		addGeneralExperts(suggestionBank["general"], &suggestions, seen, patterns, true)
		// Add triggered experts
		addGeneralExperts(suggestionBank["general"], &suggestions, seen, patterns, false)
	}

	if intention == "writing" {
		addN(suggestionBank["writing"], maxPerCategory)
	}

	if intention == "business" {
		addN(suggestionBank["business"], maxPerCategory)
		// Also add relevant general experts (Jason Fried is core)
		addGeneralExperts(suggestionBank["general"], &suggestions, seen, patterns, true)
	}

	// Phase 4: Cap at maxTotal
	if len(suggestions) > maxTotal {
		suggestions = suggestions[:maxTotal]
	}

	// Phase 5: Pre-select first N experts
	for i := range suggestions {
		suggestions[i].Selected = i < preSelectCount
	}

	return suggestions
}

// addGeneralExperts adds general experts based on core flag or trigger matching
func addGeneralExperts(experts []expert.Expert, suggestions *[]ExpertSuggestion, seen map[string]bool, patterns map[string]bool, coreOnly bool) {
	for _, e := range experts {
		if len(*suggestions) >= maxTotal {
			break
		}
		if seen[e.ID] {
			continue
		}

		shouldAdd := false
		if coreOnly {
			shouldAdd = e.Core
		} else {
			// Check if any trigger matches detected patterns
			for _, trigger := range e.Triggers {
				if patterns[trigger] {
					shouldAdd = true
					break
				}
			}
		}

		if shouldAdd {
			seen[e.ID] = true
			*suggestions = append(*suggestions, ExpertSuggestion{Expert: e})
		}
	}
}

// buildPatternSet creates a set of detected patterns for trigger matching
func buildPatternSet(d *detect.Detection) map[string]bool {
	patterns := make(map[string]bool)

	// Testing
	if len(d.Testing) > 0 {
		patterns["testing"] = true
	}

	// Map detected patterns to trigger keywords
	for _, p := range d.Patterns {
		switch p {
		case "GitHub Actions":
			patterns["ci-cd"] = true
		case "Docker":
			patterns["docker"] = true
		case "Monorepo":
			patterns["monorepo"] = true
		case "API":
			patterns["api"] = true
		case "MVC":
			patterns["mvc"] = true
		case "Microservices":
			patterns["microservices"] = true
		case "Background Jobs":
			patterns["background-jobs"] = true
		}
	}

	// Framework-based triggers
	for _, fw := range d.Frameworks {
		switch fw.Name {
		case "Rails", "Django", "Phoenix":
			patterns["mvc"] = true
			patterns["full-stack"] = true
		case "Next.js", "React", "Vue":
			patterns["frontend"] = true
		case "Express", "Gin", "FastAPI":
			patterns["api"] = true
		}
	}

	return patterns
}

// hasFramework checks if a framework is detected
func hasFramework(d *detect.Detection, name string) bool {
	for _, fw := range d.Frameworks {
		if fw.Name == name {
			return true
		}
	}
	return false
}

// selectExpertsWithSources shows a selection UI that groups by source
func selectExpertsWithSources(reader *bufio.Reader, suggestions []ExpertSuggestion, customCount, installedCount int) []expert.Expert {
	for {
		// Show custom section
		if customCount > 0 {
			fmt.Println("YOUR COUNCIL [custom]")
			for i := 0; i < customCount; i++ {
				s := suggestions[i]
				mark := "[x]"
				if !s.Selected {
					mark = "[ ]"
				}
				fmt.Printf("  %d. %s %s - %s\n", i+1, mark, s.Expert.Name, s.Expert.Focus)
			}
			fmt.Println()
		}

		// Show installed section
		if installedCount > 0 {
			fmt.Println("INSTALLED")
			start := customCount
			for i := start; i < start+installedCount; i++ {
				s := suggestions[i]
				mark := "[x]"
				if !s.Selected {
					mark = "[ ]"
				}
				source := strings.TrimPrefix(s.Source, "installed:")
				fmt.Printf("  %d. %s %s - %s [%s]\n", i+1, mark, s.Expert.Name, s.Expert.Focus, source)
			}
			fmt.Println()
		}

		// Show built-in section
		builtinStart := customCount + installedCount
		if len(suggestions) > builtinStart {
			fmt.Println("RECOMMENDED FOR THIS PROJECT")
			for i := builtinStart; i < len(suggestions); i++ {
				s := suggestions[i]
				mark := "[x]"
				if !s.Selected {
					mark = "[ ]"
				}
				fmt.Printf("  %d. %s %s - %s\n", i+1, mark, s.Expert.Name, s.Expert.Focus)
			}
			fmt.Println()
		}

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
