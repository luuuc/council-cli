package cmd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/luuuc/council-cli/internal/config"
	"github.com/luuuc/council-cli/internal/creator"
	"github.com/spf13/cobra"
)

var (
	creatorNewInterview bool
	creatorNewFrom      string
)

var (
	publishGitHub  bool
	publishPrivate bool
	publishName    string
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

	creatorPublishCmd.Flags().BoolVar(&publishGitHub, "github", false, "Publish to GitHub using gh CLI")
	creatorPublishCmd.Flags().BoolVar(&publishPrivate, "private", false, "Create private repository")
	creatorPublishCmd.Flags().StringVar(&publishName, "name", "my-council", "Repository name")
}

var creatorCmd = &cobra.Command{
	Use:   "creator",
	Short: "Manage your personal council of custom personas",
	Long: `Create and manage custom expert personas in your personal council.

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

		personas, err := creator.List()
		if err != nil {
			return err
		}

		if len(personas) == 0 {
			fmt.Println("No custom personas yet.")
			fmt.Println()
			fmt.Println("Create one with:")
			fmt.Println("  council creator new")
			fmt.Println("  council creator new --interview   (AI-assisted)")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tFOCUS\tPRIORITY")
		for _, p := range personas {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.ID, p.Name, p.Focus, p.Priority)
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

		p, err := creator.Load(args[0])
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("persona '%s' not found", args[0])
			}
			return err
		}

		fmt.Printf("ID:       %s\n", p.ID)
		fmt.Printf("Name:     %s\n", p.Name)
		fmt.Printf("Focus:    %s\n", p.Focus)
		fmt.Printf("Category: %s\n", p.Category)
		fmt.Printf("Priority: %s\n", p.Priority)

		if len(p.Triggers) > 0 {
			fmt.Println("\nTriggers:")
			for _, t := range p.Triggers {
				fmt.Printf("  - %s\n", t)
			}
		}

		if p.Philosophy != "" {
			fmt.Printf("\nPhilosophy:\n  %s\n", p.Philosophy)
		}

		if len(p.Principles) > 0 {
			fmt.Println("\nPrinciples:")
			for _, pr := range p.Principles {
				fmt.Printf("  - %s\n", pr)
			}
		}

		if len(p.RedFlags) > 0 {
			fmt.Println("\nRed Flags:")
			for _, rf := range p.RedFlags {
				fmt.Printf("  - %s\n", rf)
			}
		}

		if path, err := p.Path(); err == nil {
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
		p, err := creator.Load(id)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("persona '%s' not found", id)
			}
			return err
		}

		if !Confirm(fmt.Sprintf("Delete persona '%s'?", p.Name)) {
			fmt.Println("Cancelled.")
			return nil
		}

		if err := creator.Delete(id); err != nil {
			return err
		}

		fmt.Printf("Deleted %s\n", p.Name)
		return nil
	},
}

var creatorNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new custom persona",
	Long: `Creates a new persona through an interactive wizard.

Use --interview for AI-assisted creation where you describe someone
you value and the AI generates a persona from your description.

Use --from to fork an existing persona as a starting point.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !creator.Initialized() {
			return fmt.Errorf("personal council not initialized: run 'council creator init' first")
		}

		if creatorNewInterview {
			return runInterviewMode()
		}

		if creatorNewFrom != "" {
			return runForkMode(creatorNewFrom)
		}

		return runInteractiveWizard()
	},
}

func runInteractiveWizard() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Create New Persona")
	fmt.Println("==================")
	fmt.Println()

	// Name (required)
	name := creator.Prompt(reader, "Name:")
	if name == "" {
		return fmt.Errorf("name is required")
	}

	// Generate ID from name
	id := creator.ToID(name)
	id = creator.PromptDefault(reader, "ID:", id)

	// Check if exists
	if creator.Exists(id) {
		return fmt.Errorf("persona '%s' already exists", id)
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

	// Create persona
	p := &creator.Persona{
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
	if err := p.SaveAndCommit(true); err != nil {
		return err
	}

	path, err := p.Path()
	if err != nil {
		return fmt.Errorf("failed to get persona path: %w", err)
	}
	fmt.Println()
	fmt.Printf("Created %s\n", p.Name)
	fmt.Printf("File: %s\n", path)

	return nil
}

func runForkMode(fromID string) error {
	// Try to load from custom personas first
	var source *creator.Persona
	var err error

	source, err = creator.Load(fromID)
	if err != nil {
		// Try to find in built-in personas
		source, err = findBuiltinPersona(fromID)
		if err != nil {
			return fmt.Errorf("persona '%s' not found in custom or built-in personas", fromID)
		}
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Fork '%s' as starting point\n", source.Name)
	fmt.Println()

	// Prompt for new name
	name := creator.PromptDefault(reader, "New name:", source.Name+" (Custom)")
	id := creator.ToID(name)
	id = creator.PromptDefault(reader, "New ID:", id)

	if creator.Exists(id) {
		return fmt.Errorf("persona '%s' already exists", id)
	}

	// Create new persona based on source
	p := &creator.Persona{
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
	if err := p.SaveAndCommit(true); err != nil {
		return err
	}

	path, err := p.Path()
	if err != nil {
		return fmt.Errorf("failed to get persona path: %w", err)
	}
	fmt.Println()
	fmt.Printf("Created %s (forked from %s)\n", p.Name, source.Name)
	fmt.Printf("File: %s\n", path)

	// Offer to edit
	fmt.Println()
	if Confirm("Open in editor to customize?") {
		return openInEditor(path)
	}

	return nil
}

func runInterviewMode() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Interview Mode")
	fmt.Println("==============")
	fmt.Println()
	fmt.Println("Tell me about someone whose feedback you value.")
	fmt.Println("This could be a mentor, tech lead, author, or even a historical figure.")
	fmt.Println("Describe how they think, what they prioritize, and how they give feedback.")
	fmt.Println()
	fmt.Println("(Enter your description, then press Enter twice to finish)")
	fmt.Println()

	// Collect multi-line description
	var lines []string
	emptyCount := 0
	for {
		line, _ := reader.ReadString('\n')
		line = line[:len(line)-1] // Remove trailing newline
		if line == "" {
			emptyCount++
			if emptyCount >= 1 {
				break
			}
		} else {
			emptyCount = 0
			lines = append(lines, line)
		}
	}

	if len(lines) == 0 {
		return fmt.Errorf("no description provided")
	}

	description := ""
	for i, l := range lines {
		if i > 0 {
			description += "\n"
		}
		description += l
	}

	fmt.Println()
	fmt.Println("Generating persona from your description...")
	fmt.Println()

	// Generate persona using AI
	persona, err := generatePersonaFromDescription(description)
	if err != nil {
		return fmt.Errorf("failed to generate persona: %w", err)
	}

	// Display generated persona
	displayPersonaPreview(persona)

	// Accept/Edit/Regenerate loop
	for {
		fmt.Println()
		fmt.Print("Accept, Edit, or Regenerate? [a/e/r]: ")
		input, _ := reader.ReadString('\n')
		input = string([]byte(input[:1]))

		switch input {
		case "a", "A", "\n":
			// Accept - prompt for ID and save
			fmt.Println()
			suggestedID := creator.ToID(persona.Name)
			id := creator.PromptDefault(reader, "ID:", suggestedID)

			if creator.Exists(id) {
				return fmt.Errorf("persona '%s' already exists", id)
			}

			persona.ID = id
			if err := persona.SaveAndCommit(true); err != nil {
				return err
			}

			path, err := persona.Path()
			if err != nil {
				return fmt.Errorf("failed to get persona path: %w", err)
			}
			fmt.Println()
			fmt.Printf("Created %s\n", persona.Name)
			fmt.Printf("File: %s\n", path)
			return nil

		case "e", "E":
			// Edit in $EDITOR
			tmpfile, err := os.CreateTemp("", "council-interview-*.md")
			if err != nil {
				return fmt.Errorf("failed to create temp file: %w", err)
			}
			defer os.Remove(tmpfile.Name())

			// Write current persona to temp file
			content := formatPersonaForEdit(persona)
			if _, err := tmpfile.WriteString(content); err != nil {
				return fmt.Errorf("failed to write temp file: %w", err)
			}
			tmpfile.Close()

			// Open editor
			if err := openInEditor(tmpfile.Name()); err != nil {
				return err
			}

			// Parse edited content
			data, err := os.ReadFile(tmpfile.Name())
			if err != nil {
				return fmt.Errorf("failed to read temp file: %w", err)
			}

			edited, err := creator.Parse(data)
			if err != nil {
				fmt.Printf("Error parsing edited file: %v\n", err)
				fmt.Println("Please fix the formatting and try again.")
				continue
			}

			persona = edited
			displayPersonaPreview(persona)

		case "r", "R":
			// Regenerate
			fmt.Println()
			fmt.Println("Regenerating...")
			fmt.Println()

			persona, err = generatePersonaFromDescription(description)
			if err != nil {
				return fmt.Errorf("failed to regenerate: %w", err)
			}
			displayPersonaPreview(persona)

		default:
			fmt.Println("Invalid choice. Enter 'a' to accept, 'e' to edit, or 'r' to regenerate.")
		}
	}
}

func generatePersonaFromDescription(description string) (*creator.Persona, error) {
	// Load config for AI command
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w\nHint: run 'council init' first", err)
	}

	if cfg.AI.Command == "" {
		return nil, fmt.Errorf("no AI command configured\n\nSet in .council/config.yaml:\n  ai:\n    command: \"claude\"")
	}

	// Check if command exists
	if _, err := exec.LookPath(cfg.AI.Command); err != nil {
		return nil, fmt.Errorf("AI command '%s' not found\n\nInstall it or configure a different command", cfg.AI.Command)
	}

	// Generate prompt
	prompt := fmt.Sprintf(`Based on this description of someone the user values for feedback:

"""
%s
"""

Generate a persona file in YAML frontmatter format. The persona should capture this person's essence for code review and technical guidance.

Return ONLY valid YAML in this exact format (no markdown code blocks, no explanation):

---
id: suggested-id
name: A Memorable Name
focus: One-line description of expertise (max 60 chars)
category: custom
priority: normal
triggers:
  - keyword1
  - keyword2
  - keyword3
philosophy: |
  2-3 sentences capturing their worldview and approach.
principles:
  - First core belief or practice
  - Second core belief or practice
  - Third core belief or practice
red_flags:
  - Pattern they would call out
  - Another pattern they'd warn about
  - Third warning sign
---`, description)

	// Execute AI command
	timeout := cfg.AI.Timeout
	if timeout == 0 {
		timeout = 60
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	args := append(cfg.AI.Args, "-p", prompt)
	cmd := exec.CommandContext(ctx, cfg.AI.Command, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("AI command timed out after %d seconds", timeout)
		}
		return nil, fmt.Errorf("AI command failed: %w\n%s", err, stderr.String())
	}

	// Parse response
	response := stdout.String()

	// Try to extract YAML if wrapped in code blocks
	if idx := findYAMLStart(response); idx >= 0 {
		response = response[idx:]
	}

	persona, err := creator.Parse([]byte(response))
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w\n\nRaw response:\n%s", err, stdout.String())
	}

	return persona, nil
}

func findYAMLStart(s string) int {
	// Look for --- at start of line
	for i := 0; i < len(s); i++ {
		if i == 0 || s[i-1] == '\n' {
			if i+3 <= len(s) && s[i:i+3] == "---" {
				return i
			}
		}
	}
	return -1
}

func displayPersonaPreview(p *creator.Persona) {
	fmt.Println("┌─────────────────────────────────────────────────────┐")
	fmt.Printf("│ Name: %-45s │\n", truncate(p.Name, 45))
	fmt.Printf("│ Focus: %-44s │\n", truncate(p.Focus, 44))
	fmt.Println("│                                                     │")

	if p.Philosophy != "" {
		fmt.Println("│ Philosophy:                                         │")
		for _, line := range wrapText(p.Philosophy, 49) {
			fmt.Printf("│   %-48s │\n", line)
		}
	}

	if len(p.Principles) > 0 {
		fmt.Println("│                                                     │")
		fmt.Println("│ Principles:                                         │")
		for _, pr := range p.Principles {
			fmt.Printf("│   • %-46s │\n", truncate(pr, 46))
		}
	}

	if len(p.RedFlags) > 0 {
		fmt.Println("│                                                     │")
		fmt.Println("│ Red Flags:                                          │")
		for _, rf := range p.RedFlags {
			fmt.Printf("│   • %-46s │\n", truncate(rf, 46))
		}
	}

	fmt.Println("└─────────────────────────────────────────────────────┘")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func wrapText(s string, width int) []string {
	var lines []string
	words := splitWords(s)
	var current string

	for _, word := range words {
		if current == "" {
			current = word
		} else if len(current)+1+len(word) <= width {
			current += " " + word
		} else {
			lines = append(lines, current)
			current = word
		}
	}

	if current != "" {
		lines = append(lines, current)
	}

	return lines
}

func splitWords(s string) []string {
	var words []string
	word := ""
	for _, r := range s {
		if r == ' ' || r == '\n' || r == '\t' {
			if word != "" {
				words = append(words, word)
				word = ""
			}
		} else {
			word += string(r)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}

func formatPersonaForEdit(p *creator.Persona) string {
	var buf bytes.Buffer

	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("id: %s\n", p.ID))
	buf.WriteString(fmt.Sprintf("name: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("focus: %s\n", p.Focus))
	buf.WriteString(fmt.Sprintf("category: %s\n", p.Category))
	buf.WriteString(fmt.Sprintf("priority: %s\n", p.Priority))

	if len(p.Triggers) > 0 {
		buf.WriteString("triggers:\n")
		for _, t := range p.Triggers {
			buf.WriteString(fmt.Sprintf("  - %s\n", t))
		}
	}

	if p.Philosophy != "" {
		buf.WriteString("philosophy: |\n")
		for _, line := range wrapText(p.Philosophy, 70) {
			buf.WriteString(fmt.Sprintf("  %s\n", line))
		}
	}

	if len(p.Principles) > 0 {
		buf.WriteString("principles:\n")
		for _, pr := range p.Principles {
			buf.WriteString(fmt.Sprintf("  - %s\n", pr))
		}
	}

	if len(p.RedFlags) > 0 {
		buf.WriteString("red_flags:\n")
		for _, rf := range p.RedFlags {
			buf.WriteString(fmt.Sprintf("  - %s\n", rf))
		}
	}

	buf.WriteString("---\n")

	return buf.String()
}

func findBuiltinPersona(id string) (*creator.Persona, error) {
	// Search in suggestionBank
	for _, experts := range suggestionBank {
		for _, e := range experts {
			if e.ID == id {
				return &creator.Persona{
					ID:         e.ID,
					Name:       e.Name,
					Focus:      e.Focus,
					Philosophy: e.Philosophy,
					Principles: e.Principles,
					RedFlags:   e.RedFlags,
					Triggers:   e.Triggers,
					Priority:   "normal",
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("not found")
}

var creatorEditCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit a custom persona",
	Long:  `Opens the persona file in your $EDITOR. Changes are auto-committed.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !creator.Initialized() {
			return fmt.Errorf("personal council not initialized: run 'council creator init' first")
		}

		id := args[0]

		// Check exists
		p, err := creator.Load(id)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("persona '%s' not found", id)
			}
			return err
		}

		path, err := p.Path()
		if err != nil {
			return fmt.Errorf("failed to get persona path: %w", err)
		}

		// Open in editor
		if err := openInEditor(path); err != nil {
			return err
		}

		// Reload to verify it's still valid
		p, err = creator.Load(id)
		if err != nil {
			return fmt.Errorf("file has errors after editing: %w", err)
		}

		// Commit changes
		myCouncil, err := creator.MyCouncilPath()
		if err != nil {
			return fmt.Errorf("failed to get my-council path: %w", err)
		}
		repo := creator.NewGitRepo(myCouncil)
		if err := repo.AddAndCommit(id+".md", fmt.Sprintf("Update persona: %s", p.Name)); err != nil {
			return err
		}

		fmt.Printf("Updated %s\n", p.Name)
		return nil
	},
}

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

var creatorRemoteCmd = &cobra.Command{
	Use:   "remote [add origin <url>]",
	Short: "Manage git remote",
	Long: `View or configure the git remote for your personal council.

Examples:
  council creator remote                    # Show current remote
  council creator remote add origin <url>   # Add remote`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !creator.Initialized() {
			return fmt.Errorf("personal council not initialized: run 'council creator init' first")
		}

		myCouncil, err := creator.MyCouncilPath()
		if err != nil {
			return err
		}

		repo := creator.NewGitRepo(myCouncil)

		// No args - show current remote
		if len(args) == 0 {
			url, err := repo.RemoteGet("origin")
			if err != nil {
				fmt.Println("No remote configured")
				fmt.Println()
				fmt.Println("Add one with:")
				fmt.Println("  council creator remote add origin <url>")
				return nil
			}
			fmt.Printf("origin  %s\n", url)
			return nil
		}

		// "add origin <url>"
		if len(args) >= 3 && args[0] == "add" {
			name := args[1]
			url := args[2]

			if repo.RemoteExists(name) {
				return fmt.Errorf("remote '%s' already exists", name)
			}

			if err := repo.RemoteAdd(name, url); err != nil {
				return err
			}

			fmt.Printf("Added remote '%s': %s\n", name, url)
			return nil
		}

		return fmt.Errorf("usage: council creator remote [add <name> <url>]")
	},
}

var creatorPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push changes to remote",
	Long:  `Pushes your personal council to the configured remote repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !creator.Initialized() {
			return fmt.Errorf("personal council not initialized: run 'council creator init' first")
		}

		myCouncil, err := creator.MyCouncilPath()
		if err != nil {
			return err
		}

		repo := creator.NewGitRepo(myCouncil)

		// Check if remote exists
		if !repo.RemoteExists("origin") {
			return fmt.Errorf("no remote configured\n\nAdd one with:\n  council creator remote add origin <url>")
		}

		fmt.Println("Pushing to origin...")
		if err := repo.Push(); err != nil {
			return err
		}

		fmt.Println("Done")
		return nil
	},
}

var creatorPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull changes from remote",
	Long:  `Pulls updates from the configured remote repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !creator.Initialized() {
			return fmt.Errorf("personal council not initialized: run 'council creator init' first")
		}

		myCouncil, err := creator.MyCouncilPath()
		if err != nil {
			return err
		}

		repo := creator.NewGitRepo(myCouncil)

		// Check if remote exists
		if !repo.RemoteExists("origin") {
			return fmt.Errorf("no remote configured\n\nAdd one with:\n  council creator remote add origin <url>")
		}

		fmt.Println("Pulling from origin...")
		if err := repo.Pull(); err != nil {
			return err
		}

		fmt.Println("Done")
		return nil
	},
}

var creatorPublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish your personal council to GitHub",
	Long: `Creates a GitHub repository and pushes your personal council.

Requires the GitHub CLI (gh) to be installed and authenticated.

Examples:
  council creator publish --github           # Create public repo on GitHub
  council creator publish --github --private # Create private repo on GitHub`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !creator.Initialized() {
			return fmt.Errorf("personal council not initialized: run 'council creator init' first")
		}

		myCouncil, err := creator.MyCouncilPath()
		if err != nil {
			return err
		}

		repo := creator.NewGitRepo(myCouncil)

		// Check if remote already exists
		if repo.RemoteExists("origin") {
			url, _ := repo.RemoteGet("origin")
			fmt.Printf("Remote already configured: %s\n", url)
			fmt.Println()
			fmt.Println("To push changes:")
			fmt.Println("  council creator push")
			return nil
		}

		if !publishGitHub {
			return fmt.Errorf("specify --github to publish")
		}

		return publishToGitHub(myCouncil, repo)
	},
}

func publishToGitHub(myCouncil string, repo *creator.GitRepo) error {
	// Check if gh CLI is installed
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("GitHub CLI (gh) not found\n\nInstall it:\n  brew install gh      # macOS\n  sudo apt install gh  # Ubuntu\n\nThen authenticate:\n  gh auth login")
	}

	// Check if authenticated
	cmd := exec.Command("gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitHub CLI not authenticated\n\nRun: gh auth login")
	}

	// Build gh repo create command
	visibility := "--public"
	if publishPrivate {
		visibility = "--private"
	}

	fmt.Printf("Creating %s repository '%s'...\n", visibility[2:], publishName)

	cmd = exec.Command("gh", "repo", "create", publishName, visibility, "--source=.", "--push")
	cmd.Dir = myCouncil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	fmt.Println()
	fmt.Println("Published! Your council is now available at:")
	fmt.Printf("  https://github.com/<username>/%s\n", publishName)

	return nil
}
