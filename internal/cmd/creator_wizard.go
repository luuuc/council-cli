package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/luuuc/council-cli/internal/creator"
	"github.com/luuuc/council-cli/internal/expert"
)

// runInteractiveWizard prompts the user to create a new expert step by step.
func runInteractiveWizard() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Create New Expert")
	fmt.Println("=================")
	fmt.Println()

	// Name (required)
	name := creator.Prompt(reader, "Name:")
	if name == "" {
		return fmt.Errorf("name is required")
	}

	// Generate ID from name
	id := expert.ToID(name)
	id = creator.PromptDefault(reader, "ID:", id)

	// Check if exists
	if creator.Exists(id) {
		return fmt.Errorf("expert '%s' already exists", id)
	}

	// Focus (required)
	focus := creator.Prompt(reader, "Focus (one line):")
	if focus == "" {
		return fmt.Errorf("focus is required")
	}

	// Priority
	fmt.Println()
	priority := creator.PromptSelect(reader, "Priority (how often to suggest):", []string{
		"normal - Standard trigger matching",
		"high - Lower trigger threshold",
		"always - Always suggest, pre-selected",
	}, 0)

	// Extract just the priority value
	switch {
	case strings.HasPrefix(priority, "always"):
		priority = "always"
	case strings.HasPrefix(priority, "high"):
		priority = "high"
	default:
		priority = "normal"
	}

	// Philosophy (optional, multiline)
	fmt.Println()
	fmt.Print("Philosophy (optional) - (e)dit in editor or (s)kip? [e/s]: ")
	input, _ := reader.ReadString('\n')
	input = string([]byte(input)[0]) // Get first character

	var philosophy string
	if input == "e" || input == "E" || input == "\n" {
		var err error
		philosophy, err = creator.PromptEditor("", "Philosophy - describe this persona's worldview")
		if err != nil {
			fmt.Printf("Warning: couldn't open editor: %v\n", err)
		}
	}

	// Principles
	fmt.Println()
	principles := creator.PromptMultiline(reader, "Principles")

	// Red Flags
	fmt.Println()
	redFlags := creator.PromptMultiline(reader, "Red Flags (patterns to watch for)")

	// Triggers
	fmt.Println()
	triggers := creator.PromptMultiline(reader, "Triggers (keywords to suggest this persona)")

	// Create expert
	e := &expert.Expert{
		ID:         id,
		Name:       name,
		Focus:      focus,
		Category:   "custom",
		Priority:   priority,
		Philosophy: philosophy,
		Principles: principles,
		RedFlags:   redFlags,
		Triggers:   triggers,
	}

	// Save and commit
	if err := creator.SaveAndCommit(e, true); err != nil {
		return err
	}

	path, err := creator.ExpertPath(e.ID)
	if err != nil {
		return fmt.Errorf("failed to get expert path: %w", err)
	}
	fmt.Println()
	fmt.Printf("Created %s\n", e.Name)
	fmt.Printf("File: %s\n", path)

	return nil
}

// runForkMode creates a new expert based on an existing one.
func runForkMode(fromID string) error {
	// Try to load from custom experts first
	var source *expert.Expert
	var err error

	source, err = creator.Load(fromID)
	if err != nil {
		// Try to find in built-in experts
		source, err = findBuiltinExpert(fromID)
		if err != nil {
			return fmt.Errorf("expert '%s' not found in custom or built-in experts", fromID)
		}
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Fork '%s' as starting point\n", source.Name)
	fmt.Println()

	// Prompt for new name
	name := creator.PromptDefault(reader, "New name:", source.Name+" (Custom)")
	id := expert.ToID(name)
	id = creator.PromptDefault(reader, "New ID:", id)

	if creator.Exists(id) {
		return fmt.Errorf("expert '%s' already exists", id)
	}

	// Create new expert based on source
	e := &expert.Expert{
		ID:         id,
		Name:       name,
		Focus:      source.Focus,
		Category:   "custom",
		Priority:   source.Priority,
		Philosophy: source.Philosophy,
		Principles: source.Principles,
		RedFlags:   source.RedFlags,
		Triggers:   source.Triggers,
	}

	// Save and commit
	if err := creator.SaveAndCommit(e, true); err != nil {
		return err
	}

	path, err := creator.ExpertPath(e.ID)
	if err != nil {
		return fmt.Errorf("failed to get expert path: %w", err)
	}
	fmt.Println()
	fmt.Printf("Created %s (forked from %s)\n", e.Name, source.Name)
	fmt.Printf("File: %s\n", path)

	// Offer to edit
	fmt.Println()
	if Confirm("Open in editor to customize?") {
		return openInEditor(path)
	}

	return nil
}
