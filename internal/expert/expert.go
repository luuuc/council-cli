package expert

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/luuuc/council-cli/internal/config"
	"gopkg.in/yaml.v3"
)

// Expert represents an expert persona
type Expert struct {
	ID         string   `yaml:"id"`
	Name       string   `yaml:"name"`
	Focus      string   `yaml:"focus"`
	Triggers   Triggers `yaml:"triggers,omitempty"`
	Philosophy string   `yaml:"philosophy,omitempty"`
	Principles []string `yaml:"principles,omitempty"`
	RedFlags   []string `yaml:"red_flags,omitempty"`

	// Body is the markdown content after frontmatter
	Body string `yaml:"-"`
}

// Triggers defines when an expert should be consulted
type Triggers struct {
	Paths    []string `yaml:"paths,omitempty"`
	Keywords []string `yaml:"keywords,omitempty"`
}

// ExpertSuggestions is the expected AI response format
type ExpertSuggestions struct {
	Experts []Expert `yaml:"experts"`
}

// Path returns the file path for this expert
func (e *Expert) Path() string {
	return config.Path(config.ExpertsDir, e.ID+".md")
}

// Save writes the expert to disk
func (e *Expert) Save() error {
	// Generate body if empty
	if e.Body == "" {
		e.Body = e.generateBody()
	}

	// Generate frontmatter
	fm, err := yaml.Marshal(e)
	if err != nil {
		return fmt.Errorf("failed to marshal expert: %w", err)
	}

	// Combine frontmatter and body
	content := fmt.Sprintf("---\n%s---\n\n%s", string(fm), e.Body)

	// Ensure directory exists
	dir := filepath.Dir(e.Path())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(e.Path(), []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write expert file: %w", err)
	}

	return nil
}

func (e *Expert) generateBody() string {
	tmpl := `# {{.Name}} - {{.Focus}}

You are channeling {{.Name}}, known for expertise in {{.Focus}}.

{{if .Philosophy}}## Philosophy

{{.Philosophy}}
{{end}}
{{if .Principles}}## Principles

{{range .Principles}}- {{.}}
{{end}}
{{end}}
{{if .RedFlags}}## Red Flags

Watch for these patterns:
{{range .RedFlags}}- {{.}}
{{end}}
{{end}}
## Review Style

When reviewing code, focus on your area of expertise. Be direct and specific.
Explain your reasoning. Suggest concrete improvements.
`
	t, _ := template.New("body").Parse(tmpl)
	var buf bytes.Buffer
	t.Execute(&buf, e)
	return strings.TrimSpace(buf.String())
}

// Load reads an expert from disk
func Load(id string) (*Expert, error) {
	path := config.Path(config.ExpertsDir, id+".md")
	return LoadFile(path)
}

// LoadFile reads an expert from a specific file
func LoadFile(path string) (*Expert, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return Parse(data)
}

// Parse parses expert markdown with frontmatter
func Parse(data []byte) (*Expert, error) {
	content := string(data)

	// Split frontmatter and body
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("missing frontmatter")
	}

	parts := strings.SplitN(content[3:], "---", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid frontmatter format")
	}

	frontmatter := strings.TrimSpace(parts[0])
	body := strings.TrimSpace(parts[1])

	var e Expert
	if err := yaml.Unmarshal([]byte(frontmatter), &e); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	e.Body = body
	return &e, nil
}

// List returns all experts in the council
func List() ([]*Expert, error) {
	dir := config.Path(config.ExpertsDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var experts []*Expert
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		if entry.Name() == ".gitkeep" {
			continue
		}

		e, err := LoadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue // Skip invalid files
		}
		experts = append(experts, e)
	}

	return experts, nil
}

// Delete removes an expert from the council
func Delete(id string) error {
	path := config.Path(config.ExpertsDir, id+".md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("expert '%s' not found", id)
	}
	return os.Remove(path)
}

// Exists checks if an expert exists
func Exists(id string) bool {
	path := config.Path(config.ExpertsDir, id+".md")
	_, err := os.Stat(path)
	return err == nil
}

// ToID converts a name to a kebab-case ID
func ToID(name string) string {
	// Convert to lowercase
	id := strings.ToLower(name)

	// Replace spaces and special chars with hyphens
	re := regexp.MustCompile(`[^a-z0-9]+`)
	id = re.ReplaceAllString(id, "-")

	// Remove leading/trailing hyphens
	id = strings.Trim(id, "-")

	return id
}

// ParseAIResponse parses YAML response from AI into experts
func ParseAIResponse(data []byte) ([]Expert, error) {
	// Try to extract YAML from markdown code block
	content := string(data)
	if idx := strings.Index(content, "```yaml"); idx >= 0 {
		content = content[idx+7:]
		if end := strings.Index(content, "```"); end >= 0 {
			content = content[:end]
		}
	} else if idx := strings.Index(content, "```"); idx >= 0 {
		content = content[idx+3:]
		if end := strings.Index(content, "```"); end >= 0 {
			content = content[:end]
		}
	}

	var suggestions ExpertSuggestions
	if err := yaml.Unmarshal([]byte(content), &suggestions); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return suggestions.Experts, nil
}

// Confirm asks user for confirmation
func Confirm(prompt string) bool {
	fmt.Print(prompt + " [Y/n] ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "" || response == "y" || response == "yes"
}
