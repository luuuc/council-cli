package cmd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/luuuc/council-cli/internal/config"
	"github.com/luuuc/council-cli/internal/creator"
	"github.com/luuuc/council-cli/internal/expert"
)

// runInterviewMode uses AI to generate an expert from a description.
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
	fmt.Println("Generating expert from your description...")
	fmt.Println()

	// Generate expert using AI
	exp, err := generateExpertFromDescription(description)
	if err != nil {
		return fmt.Errorf("failed to generate expert: %w", err)
	}

	// Display generated expert
	displayExpertPreview(exp)

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
			suggestedID := expert.ToID(exp.Name)
			id := creator.PromptDefault(reader, "ID:", suggestedID)

			if creator.Exists(id) {
				return fmt.Errorf("expert '%s' already exists", id)
			}

			exp.ID = id
			if err := creator.SaveAndCommit(exp, true); err != nil {
				return err
			}

			path, err := creator.ExpertPath(exp.ID)
			if err != nil {
				return fmt.Errorf("failed to get expert path: %w", err)
			}
			fmt.Println()
			fmt.Printf("Created %s\n", exp.Name)
			fmt.Printf("File: %s\n", path)
			return nil

		case "e", "E":
			// Edit in $EDITOR
			tmpfile, err := os.CreateTemp("", "council-interview-*.md")
			if err != nil {
				return fmt.Errorf("failed to create temp file: %w", err)
			}
			defer os.Remove(tmpfile.Name())

			// Write current expert to temp file
			content := formatExpertForEdit(exp)
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

			exp = edited
			displayExpertPreview(exp)

		case "r", "R":
			// Regenerate
			fmt.Println()
			fmt.Println("Regenerating...")
			fmt.Println()

			exp, err = generateExpertFromDescription(description)
			if err != nil {
				return fmt.Errorf("failed to regenerate: %w", err)
			}
			displayExpertPreview(exp)

		default:
			fmt.Println("Invalid choice. Enter 'a' to accept, 'e' to edit, or 'r' to regenerate.")
		}
	}
}

// generateExpertFromDescription uses AI to create an expert from a description.
func generateExpertFromDescription(description string) (*expert.Expert, error) {
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

Generate an expert file in YAML frontmatter format. The expert should capture this person's essence for code review and technical guidance.

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

	exp, err := creator.Parse([]byte(response))
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w\n\nRaw response:\n%s", err, stdout.String())
	}

	return exp, nil
}

// findYAMLStart finds the start of YAML frontmatter in a string.
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

// displayExpertPreview shows a formatted preview of an expert.
func displayExpertPreview(e *expert.Expert) {
	fmt.Println("┌─────────────────────────────────────────────────────┐")
	fmt.Printf("│ Name: %-45s │\n", truncate(e.Name, 45))
	fmt.Printf("│ Focus: %-44s │\n", truncate(e.Focus, 44))
	fmt.Println("│                                                     │")

	if e.Philosophy != "" {
		fmt.Println("│ Philosophy:                                         │")
		for _, line := range wrapText(e.Philosophy, 49) {
			fmt.Printf("│   %-48s │\n", line)
		}
	}

	if len(e.Principles) > 0 {
		fmt.Println("│                                                     │")
		fmt.Println("│ Principles:                                         │")
		for _, pr := range e.Principles {
			fmt.Printf("│   • %-46s │\n", truncate(pr, 46))
		}
	}

	if len(e.RedFlags) > 0 {
		fmt.Println("│                                                     │")
		fmt.Println("│ Red Flags:                                          │")
		for _, rf := range e.RedFlags {
			fmt.Printf("│   • %-46s │\n", truncate(rf, 46))
		}
	}

	fmt.Println("└─────────────────────────────────────────────────────┘")
}

// formatExpertForEdit formats an expert for editing in a text editor.
func formatExpertForEdit(e *expert.Expert) string {
	var buf bytes.Buffer

	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("id: %s\n", e.ID))
	buf.WriteString(fmt.Sprintf("name: %s\n", e.Name))
	buf.WriteString(fmt.Sprintf("focus: %s\n", e.Focus))
	buf.WriteString(fmt.Sprintf("category: %s\n", e.Category))
	buf.WriteString(fmt.Sprintf("priority: %s\n", e.Priority))

	if len(e.Triggers) > 0 {
		buf.WriteString("triggers:\n")
		for _, t := range e.Triggers {
			buf.WriteString(fmt.Sprintf("  - %s\n", t))
		}
	}

	if e.Philosophy != "" {
		buf.WriteString("philosophy: |\n")
		for _, line := range wrapText(e.Philosophy, 70) {
			buf.WriteString(fmt.Sprintf("  %s\n", line))
		}
	}

	if len(e.Principles) > 0 {
		buf.WriteString("principles:\n")
		for _, pr := range e.Principles {
			buf.WriteString(fmt.Sprintf("  - %s\n", pr))
		}
	}

	if len(e.RedFlags) > 0 {
		buf.WriteString("red_flags:\n")
		for _, rf := range e.RedFlags {
			buf.WriteString(fmt.Sprintf("  - %s\n", rf))
		}
	}

	buf.WriteString("---\n")

	return buf.String()
}

// findBuiltinExpert finds an expert in the built-in suggestion bank.
func findBuiltinExpert(id string) (*expert.Expert, error) {
	// Search in suggestionBank
	for _, experts := range suggestionBank {
		for _, e := range experts {
			if e.ID == id {
				return &expert.Expert{
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

// truncate shortens a string to maxLen, adding "..." if needed.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// wrapText wraps text to a given width.
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

// splitWords splits a string into words.
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
