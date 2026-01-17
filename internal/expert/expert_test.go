package expert

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luuuc/council-cli/internal/config"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Expert
		wantErr bool
	}{
		{
			name: "valid expert with all fields",
			input: `---
id: kent-beck
name: Kent Beck
focus: Test-driven development and simple design
philosophy: Write tests first, then code.
principles:
  - Red-green-refactor
  - Simple design over complexity
red_flags:
  - Tests written after code
  - Over-engineering
---

# Kent Beck

Expert in testing.`,
			want: &Expert{
				ID:         "kent-beck",
				Name:       "Kent Beck",
				Focus:      "Test-driven development and simple design",
				Philosophy: "Write tests first, then code.",
				Principles: []string{"Red-green-refactor", "Simple design over complexity"},
				RedFlags:   []string{"Tests written after code", "Over-engineering"},
				Body:       "# Kent Beck\n\nExpert in testing.",
			},
			wantErr: false,
		},
		{
			name: "minimal expert",
			input: `---
id: minimal
name: Minimal Expert
focus: Minimal focus
---

Body content.`,
			want: &Expert{
				ID:    "minimal",
				Name:  "Minimal Expert",
				Focus: "Minimal focus",
				Body:  "Body content.",
			},
			wantErr: false,
		},
		{
			name:    "missing frontmatter",
			input:   "No frontmatter here",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid frontmatter format",
			input:   "---\nno closing delimiter",
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid yaml in frontmatter",
			input: `---
id: [invalid yaml
name: broken
---

Body.`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if got.ID != tt.want.ID {
				t.Errorf("Parse() ID = %v, want %v", got.ID, tt.want.ID)
			}
			if got.Name != tt.want.Name {
				t.Errorf("Parse() Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Focus != tt.want.Focus {
				t.Errorf("Parse() Focus = %v, want %v", got.Focus, tt.want.Focus)
			}
			if got.Philosophy != tt.want.Philosophy {
				t.Errorf("Parse() Philosophy = %v, want %v", got.Philosophy, tt.want.Philosophy)
			}
			if got.Body != tt.want.Body {
				t.Errorf("Parse() Body = %v, want %v", got.Body, tt.want.Body)
			}
			if len(got.Principles) != len(tt.want.Principles) {
				t.Errorf("Parse() Principles len = %v, want %v", len(got.Principles), len(tt.want.Principles))
			}
			if len(got.RedFlags) != len(tt.want.RedFlags) {
				t.Errorf("Parse() RedFlags len = %v, want %v", len(got.RedFlags), len(tt.want.RedFlags))
			}
		})
	}
}

func TestSave(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "council-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory and set up council structure
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	// Create the council directory structure
	_ = os.MkdirAll(config.Path(config.ExpertsDir), 0755)

	tests := []struct {
		name    string
		expert  *Expert
		wantErr bool
	}{
		{
			name: "save expert with all fields",
			expert: &Expert{
				ID:         "dhh",
				Name:       "DHH",
				Focus:      "Rails and convention over configuration",
				Philosophy: "Optimize for programmer happiness.",
				Principles: []string{"Convention over configuration", "DRY"},
				RedFlags:   []string{"Too much abstraction"},
			},
			wantErr: false,
		},
		{
			name: "save minimal expert",
			expert: &Expert{
				ID:    "minimal",
				Name:  "Minimal",
				Focus: "Testing",
			},
			wantErr: false,
		},
		{
			name: "save expert with custom body",
			expert: &Expert{
				ID:    "custom",
				Name:  "Custom",
				Focus: "Custom focus",
				Body:  "# Custom Body\n\nThis is a custom body.",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.expert.Save()
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify file was created
			path := tt.expert.Path()
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("Save() did not create file at %s", path)
				return
			}

			// Load it back and verify
			loaded, err := Load(tt.expert.ID)
			if err != nil {
				t.Errorf("Failed to load saved expert: %v", err)
				return
			}

			if loaded.ID != tt.expert.ID {
				t.Errorf("Loaded ID = %v, want %v", loaded.ID, tt.expert.ID)
			}
			if loaded.Name != tt.expert.Name {
				t.Errorf("Loaded Name = %v, want %v", loaded.Name, tt.expert.Name)
			}
			if loaded.Focus != tt.expert.Focus {
				t.Errorf("Loaded Focus = %v, want %v", loaded.Focus, tt.expert.Focus)
			}
		})
	}
}

func TestToID(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Kent Beck", "kent-beck"},
		{"DHH", "dhh"},
		{"José Valim", "jos-valim"},
		{"Rob Pike", "rob-pike"},
		{"Already-Kebab", "already-kebab"},
		{"  Spaces  Around  ", "spaces-around"},
		{"Special!@#$%Characters", "special-characters"},
		{"123Numbers", "123numbers"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToID(tt.name); got != tt.want {
				t.Errorf("ToID(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestParseAIResponse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
	}{
		{
			name: "yaml in code block",
			input: `Here are the experts:

` + "```yaml" + `
experts:
  - id: kent-beck
    name: Kent Beck
    focus: TDD
  - id: dhh
    name: DHH
    focus: Rails
` + "```" + `

That's all!`,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "yaml in generic code block",
			input: "```\nexperts:\n  - id: test\n    name: Test\n    focus: Testing\n```",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "plain yaml",
			input: `experts:
  - id: plain
    name: Plain Expert
    focus: Plain focus`,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "invalid yaml",
			input:     "not: [valid: yaml",
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "empty experts",
			input:     "experts: []",
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAIResponse([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAIResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("ParseAIResponse() returned %d experts, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestListWithWarnings(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "council-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	// Create the council directory structure
	expertsDir := config.Path(config.ExpertsDir)
	_ = os.MkdirAll(expertsDir, 0755)

	// Create a valid expert file
	validExpert := &Expert{
		ID:    "valid",
		Name:  "Valid Expert",
		Focus: "Testing",
	}
	_ = validExpert.Save()

	// Create an invalid expert file
	invalidPath := filepath.Join(expertsDir, "invalid.md")
	_ = os.WriteFile(invalidPath, []byte("not valid frontmatter"), 0644)

	result, err := ListWithWarnings()
	if err != nil {
		t.Fatalf("ListWithWarnings() error = %v", err)
	}

	if len(result.Experts) != 1 {
		t.Errorf("ListWithWarnings() got %d experts, want 1", len(result.Experts))
	}

	if len(result.Warnings) != 1 {
		t.Errorf("ListWithWarnings() got %d warnings, want 1", len(result.Warnings))
	}

	if len(result.Experts) > 0 && result.Experts[0].ID != "valid" {
		t.Errorf("ListWithWarnings() expert ID = %v, want valid", result.Experts[0].ID)
	}
}

func TestExistsAndDelete(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "council-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	// Create the council directory structure
	_ = os.MkdirAll(config.Path(config.ExpertsDir), 0755)

	// Test that non-existent expert doesn't exist
	if Exists("nonexistent") {
		t.Error("Exists() returned true for non-existent expert")
	}

	// Create an expert
	expert := &Expert{
		ID:    "test",
		Name:  "Test Expert",
		Focus: "Testing",
	}
	_ = expert.Save()

	// Test that it now exists
	if !Exists("test") {
		t.Error("Exists() returned false for existing expert")
	}

	// Delete it
	err = Delete("test")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Test that it no longer exists
	if Exists("test") {
		t.Error("Exists() returned true after deletion")
	}

	// Test deleting non-existent expert
	err = Delete("nonexistent")
	if err == nil {
		t.Error("Delete() should error for non-existent expert")
	}
}

func TestGenerateBody(t *testing.T) {
	expert := &Expert{
		ID:         "test",
		Name:       "Test Expert",
		Focus:      "Testing methodology",
		Philosophy: "Test everything.",
		Principles: []string{"Write tests first", "Keep tests simple"},
		RedFlags:   []string{"No tests", "Flaky tests"},
	}

	body := expert.generateBody()

	// Check that key elements are present
	if !strings.Contains(body, "Test Expert") {
		t.Error("generateBody() missing expert name")
	}
	if !strings.Contains(body, "Testing methodology") {
		t.Error("generateBody() missing focus")
	}
	if !strings.Contains(body, "Test everything.") {
		t.Error("generateBody() missing philosophy")
	}
	if !strings.Contains(body, "Write tests first") {
		t.Error("generateBody() missing principle")
	}
	if !strings.Contains(body, "No tests") {
		t.Error("generateBody() missing red flag")
	}
}

func TestGenerateBody_MinimalExpert(t *testing.T) {
	// Test with minimal fields - should not panic and should produce valid output
	expert := &Expert{
		ID:    "minimal",
		Name:  "Minimal",
		Focus: "Minimal focus",
	}

	body := expert.generateBody()

	if body == "" {
		t.Error("generateBody() should not return empty string for minimal expert")
	}
	if !strings.Contains(body, "Minimal") {
		t.Error("generateBody() should contain expert name")
	}
	if !strings.Contains(body, "Minimal focus") {
		t.Error("generateBody() should contain focus")
	}
	// Should not contain optional sections when they're empty
	if strings.Contains(body, "## Philosophy") {
		t.Error("generateBody() should not contain Philosophy section when empty")
	}
}

func TestGenerateBody_EmptyStrings(t *testing.T) {
	// Test with empty strings - should handle gracefully
	expert := &Expert{
		ID:         "",
		Name:       "",
		Focus:      "",
		Philosophy: "",
		Principles: []string{},
		RedFlags:   []string{},
	}

	body := expert.generateBody()

	// Should not panic and should produce some output
	if body == "" {
		t.Error("generateBody() should not return empty string")
	}
	// The template should still have the Review Style section
	if !strings.Contains(body, "Review Style") {
		t.Error("generateBody() should contain Review Style section")
	}
}

func TestGenerateBody_SpecialCharacters(t *testing.T) {
	// Test with special characters that might cause template issues
	expert := &Expert{
		ID:         "special",
		Name:       "Expert with <html> & \"quotes\"",
		Focus:      "Testing {{templates}}",
		Philosophy: "Philosophy with `backticks` and $variables",
		Principles: []string{"Principle with {{braces}}"},
		RedFlags:   []string{"Red flag with <angle> brackets"},
	}

	body := expert.generateBody()

	// Should not panic and should contain the special characters as-is
	// (text/template does not escape by default)
	if !strings.Contains(body, "<html>") {
		t.Error("generateBody() should preserve HTML-like content")
	}
	if !strings.Contains(body, "{{templates}}") {
		t.Error("generateBody() should preserve template-like content as literal text")
	}
}
