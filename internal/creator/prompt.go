package creator

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var idRegexp = regexp.MustCompile(`[^a-z0-9]+`)

// ToID converts a name to a kebab-case ID.
func ToID(name string) string {
	id := strings.ToLower(name)
	id = idRegexp.ReplaceAllString(id, "-")
	id = strings.Trim(id, "-")
	return id
}

// Prompt asks a question and returns the answer.
func Prompt(reader *bufio.Reader, question string) string {
	fmt.Print(question + " ")
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// PromptDefault asks a question with a default value.
func PromptDefault(reader *bufio.Reader, question, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s] ", question, defaultVal)
	} else {
		fmt.Print(question + " ")
	}
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}
	return input
}

// PromptMultiline prompts for multiple lines until empty line.
func PromptMultiline(reader *bufio.Reader, question string) []string {
	fmt.Println(question + " (one per line, empty line to finish):")
	var lines []string
	for {
		fmt.Print("  > ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			break
		}
		lines = append(lines, input)
	}
	return lines
}

// PromptSelect shows a selection menu and returns the chosen value.
func PromptSelect(reader *bufio.Reader, question string, options []string, defaultIndex int) string {
	fmt.Println(question)
	for i, opt := range options {
		marker := "  "
		if i == defaultIndex {
			marker = "> "
		}
		fmt.Printf("  %s%d. %s\n", marker, i+1, opt)
	}
	fmt.Printf("Choice [%d]: ", defaultIndex+1)

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return options[defaultIndex]
	}

	var idx int
	if _, err := fmt.Sscanf(input, "%d", &idx); err == nil && idx >= 1 && idx <= len(options) {
		return options[idx-1]
	}

	return options[defaultIndex]
}

// PromptEditor opens the user's $EDITOR for multiline input.
func PromptEditor(initial, header string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		// Default editors by platform
		if _, err := exec.LookPath("vim"); err == nil {
			editor = "vim"
		} else if _, err := exec.LookPath("nano"); err == nil {
			editor = "nano"
		} else if _, err := exec.LookPath("notepad"); err == nil {
			editor = "notepad"
		} else {
			return "", fmt.Errorf("no editor found: set $EDITOR environment variable")
		}
	}

	// Create temp file
	tmpfile, err := os.CreateTemp("", "council-*.md")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write header and initial content
	content := ""
	if header != "" {
		content = "# " + header + "\n# Lines starting with # will be removed\n\n"
	}
	content += initial

	if _, err := tmpfile.WriteString(content); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpfile.Close()

	// Open editor
	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor failed: %w", err)
	}

	// Read result
	data, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read temp file: %w", err)
	}

	// Remove comment lines
	lines := strings.Split(string(data), "\n")
	var result []string
	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			result = append(result, line)
		}
	}

	return strings.TrimSpace(strings.Join(result, "\n")), nil
}

// Confirm asks for yes/no confirmation.
func Confirm(prompt string) bool {
	fmt.Print(prompt + " [Y/n] ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "" || response == "y" || response == "yes"
}
