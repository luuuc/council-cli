package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/luuuc/council-cli/internal/detect"
	"github.com/luuuc/council-cli/internal/expert"
)

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

	// Language-specific experts
	for _, lang := range d.Languages {
		switch lang.Name {
		case "Go":
			suggestions = append(suggestions, goExperts()...)
		case "Ruby":
			suggestions = append(suggestions, rubyExperts()...)
		case "Python":
			suggestions = append(suggestions, pythonExperts()...)
		case "TypeScript", "JavaScript":
			suggestions = append(suggestions, jsExperts()...)
		case "Rust":
			suggestions = append(suggestions, rustExperts()...)
		}
	}

	// Add scope/simplicity experts for code projects
	if intention == "code" || intention == "everything" {
		suggestions = appendIfNotExists(suggestions, jasonFried())
		suggestions = appendIfNotExists(suggestions, dieterRams())
	}

	// Writing experts
	if intention == "writing" {
		suggestions = append(suggestions, writingExperts()...)
	}

	// Business experts
	if intention == "business" {
		suggestions = append(suggestions, businessExperts()...)
	}

	// Deduplicate
	return deduplicateSuggestions(suggestions)
}

func appendIfNotExists(suggestions []ExpertSuggestion, s ExpertSuggestion) []ExpertSuggestion {
	for _, existing := range suggestions {
		if existing.Expert.ID == s.Expert.ID {
			return suggestions
		}
	}
	return append(suggestions, s)
}

func deduplicateSuggestions(suggestions []ExpertSuggestion) []ExpertSuggestion {
	seen := make(map[string]bool)
	var result []ExpertSuggestion
	for _, s := range suggestions {
		if !seen[s.Expert.ID] {
			seen[s.Expert.ID] = true
			result = append(result, s)
		}
	}
	return result
}

// Expert suggestion banks

func goExperts() []ExpertSuggestion {
	return []ExpertSuggestion{
		{Expert: expert.Expert{
			ID:    "rob-pike",
			Name:  "Rob Pike",
			Focus: "Clarity, simplicity, and idiomatic Go",
			Philosophy: `Clarity is paramount. Programs are read far more often than written.
Go was designed to be boring on purpose - boring means predictable, maintainable.
A little copying is better than a little dependency.`,
			Principles: []string{
				"Clear is better than clever",
				"A little copying is better than a little dependency",
				"The bigger the interface, the weaker the abstraction",
				"Make the zero value useful",
			},
			RedFlags: []string{
				"Interfaces with only one implementation",
				"Clever code that requires explanation",
				"Deep package hierarchies",
			},
		}},
		{Expert: expert.Expert{
			ID:    "dave-cheney",
			Name:  "Dave Cheney",
			Focus: "Go performance, APIs, and maintainability",
			Philosophy: `Good code is not just correct today; it remains understandable years from now.
APIs, once exported, are promises. Performance is a feature, but measure first.`,
			Principles: []string{
				"APIs are forever - design them carefully",
				"Measure before optimizing",
				"Prefer small, composable packages",
				"Design for change",
			},
			RedFlags: []string{
				"Exported APIs without clear rationale",
				"Performance claims without benchmarks",
			},
		}},
	}
}

func rubyExperts() []ExpertSuggestion {
	return []ExpertSuggestion{
		{Expert: expert.Expert{
			ID:    "dhh",
			Name:  "DHH",
			Focus: "Rails doctrine and convention over configuration",
			Philosophy: `Convention over configuration frees you to focus on what matters.
The Majestic Monolith is underrated. Integrated systems beat distributed complexity.`,
			Principles: []string{
				"Convention over configuration",
				"Programmer happiness matters",
				"Majestic monolith over microservices",
				"No service objects - use models and controllers",
			},
			RedFlags: []string{
				"Service objects everywhere",
				"Premature extraction to microservices",
				"Fighting the framework",
			},
		}},
		{Expert: expert.Expert{
			ID:    "sandi-metz",
			Name:  "Sandi Metz",
			Focus: "Object-oriented design and practical refactoring",
			Philosophy: `Duplication is far cheaper than the wrong abstraction.
Small objects that do one thing lead to flexible systems.`,
			Principles: []string{
				"Prefer duplication over the wrong abstraction",
				"Small objects, small methods",
				"Depend on behavior, not data",
				"Refactor when you understand the pattern",
			},
			RedFlags: []string{
				"God objects doing everything",
				"Premature abstraction",
				"Methods longer than 5 lines (guideline)",
			},
		}},
	}
}

func pythonExperts() []ExpertSuggestion {
	return []ExpertSuggestion{
		{Expert: expert.Expert{
			ID:    "raymond-hettinger",
			Name:  "Raymond Hettinger",
			Focus: "Pythonic code and standard library mastery",
			Philosophy: `There should be one obvious way to do it. Beautiful is better than ugly.
Use the standard library. Know your data structures.`,
			Principles: []string{
				"Pythonic over clever",
				"Use built-in functions and standard library",
				"Flat is better than nested",
				"Readability counts",
			},
			RedFlags: []string{
				"Reinventing standard library features",
				"Overly nested code",
				"Java-style Python",
			},
		}},
	}
}

func jsExperts() []ExpertSuggestion {
	return []ExpertSuggestion{
		{Expert: expert.Expert{
			ID:    "dan-abramov",
			Name:  "Dan Abramov",
			Focus: "React patterns and JavaScript fundamentals",
			Philosophy: `Understand the fundamentals deeply. Don't cargo-cult patterns.
Keep components simple. State should be minimal and derived when possible.`,
			Principles: []string{
				"Understand before abstracting",
				"Minimal state, derive the rest",
				"Composition over inheritance",
				"Keep side effects at the edges",
			},
			RedFlags: []string{
				"Overusing Redux for local state",
				"Premature optimization",
				"Copy-paste from Stack Overflow without understanding",
			},
		}},
	}
}

func rustExperts() []ExpertSuggestion {
	return []ExpertSuggestion{
		{Expert: expert.Expert{
			ID:    "steve-klabnik",
			Name:  "Steve Klabnik",
			Focus: "Rust idioms and documentation",
			Philosophy: `Documentation is a feature. The compiler is your friend.
Embrace the borrow checker - it catches real bugs.`,
			Principles: []string{
				"Let the compiler help you",
				"Document public APIs thoroughly",
				"Prefer owned types for simplicity when performance allows",
				"Use the type system to prevent bugs",
			},
			RedFlags: []string{
				"Fighting the borrow checker with unsafe",
				"Missing documentation on public items",
				"Ignoring clippy warnings",
			},
		}},
	}
}

func writingExperts() []ExpertSuggestion {
	return []ExpertSuggestion{
		{Expert: expert.Expert{
			ID:    "william-zinsser",
			Name:  "William Zinsser",
			Focus: "Clarity and simplicity in non-fiction",
			Philosophy: `Writing is hard work. Simplify, then simplify again.
Every word must earn its place. Strip every sentence to its cleanest components.`,
			Principles: []string{
				"Simplify, simplify",
				"Clear thinking leads to clear writing",
				"Remove clutter ruthlessly",
				"Write for yourself first",
			},
			RedFlags: []string{
				"Unnecessary words and phrases",
				"Passive voice when active works",
				"Jargon that excludes readers",
			},
		}},
		{Expert: expert.Expert{
			ID:    "stephen-king",
			Name:  "Stephen King",
			Focus: "Practical writing craft",
			Philosophy: `The adverb is not your friend. Show, don't tell.
Write with the door closed, rewrite with the door open.`,
			Principles: []string{
				"Kill your darlings",
				"Second draft = first draft minus 10%",
				"Read a lot, write a lot",
				"Avoid adverbs, especially in dialogue",
			},
			RedFlags: []string{
				"Overwriting - too many adverbs and adjectives",
				"Telling when showing would work",
				"Passive voice sapping energy",
			},
		}},
	}
}

func businessExperts() []ExpertSuggestion {
	return []ExpertSuggestion{
		{Expert: expert.Expert{
			ID:    "paul-graham",
			Name:  "Paul Graham",
			Focus: "Startup discipline and focus",
			Philosophy: `Make something people want. Do things that don't scale.
Talk to users obsessively. Startups die from suicide, not murder.`,
			Principles: []string{
				"Make something people want",
				"Do things that don't scale",
				"Launch early, iterate fast",
				"Focus on the users who love you",
			},
			RedFlags: []string{
				"Building without talking to users",
				"Premature scaling",
				"Vanity metrics",
			},
		}},
		jasonFried(),
	}
}

func jasonFried() ExpertSuggestion {
	return ExpertSuggestion{Expert: expert.Expert{
		ID:    "jason-fried",
		Name:  "Jason Fried",
		Focus: "Scope discipline, simplicity, and shipping",
		Philosophy: `Do less, but do it well. Half a product is better than a half-assed product.
Every feature is a liability. Complexity is debt with interest.`,
		Principles: []string{
			"Do less, but do it well",
			"Half a product, not a half-assed product",
			"Ship something real, then iterate",
			"Every feature is a liability",
		},
		RedFlags: []string{
			"Feature requests without clear problem statements",
			"Building for hypothetical future needs",
			"Adding configurability instead of making decisions",
		},
	}}
}

func dieterRams() ExpertSuggestion {
	return ExpertSuggestion{Expert: expert.Expert{
		ID:    "dieter-rams",
		Name:  "Dieter Rams",
		Focus: "Design simplicity and essentialism",
		Philosophy: `Less, but better. Good design is as little design as possible.
Every element must justify its existence. Good design is honest.`,
		Principles: []string{
			"Less, but better",
			"Good design is as little design as possible",
			"Every element must earn its place",
			"Good design is honest",
		},
		RedFlags: []string{
			"Features that exist because we can",
			"Complexity without clear benefit",
			"Decoration over function",
		},
	}}
}
