package creator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Persona represents a custom persona in my-council.
// Uses the same format as built-in experts with additional priority field.
type Persona struct {
	ID         string   `yaml:"id"`
	Name       string   `yaml:"name"`
	Focus      string   `yaml:"focus"`
	Category   string   `yaml:"category,omitempty"`
	Priority   string   `yaml:"priority,omitempty"` // always, high, normal
	Triggers   []string `yaml:"triggers,omitempty"`
	Philosophy string   `yaml:"philosophy,omitempty"`
	Principles []string `yaml:"principles,omitempty"`
	RedFlags   []string `yaml:"red_flags,omitempty"`

	// Body is the markdown content after frontmatter
	Body string `yaml:"-"`

	// Source indicates where this persona came from (custom, installed:<name>)
	Source string `yaml:"-"`
}

// Init initializes the personal council directory with git.
func Init() error {
	path, err := MyCouncilPath()
	if err != nil {
		return fmt.Errorf("failed to get my-council path: %w", err)
	}

	// Check if already initialized
	if Initialized() {
		return fmt.Errorf("my-council already initialized at %s", path)
	}

	// Create directory
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Initialize git repo
	repo := NewGitRepo(path)
	if err := repo.Init(); err != nil {
		return fmt.Errorf("failed to initialize git: %w", err)
	}

	// Create README.md
	readme := generateREADME()
	readmePath := filepath.Join(path, "README.md")
	if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	// Initial commit
	if err := repo.Add("."); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}
	if err := repo.Commit("Initialize my-council"); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// generateREADME creates the README.md template.
func generateREADME() string {
	return `# My Council

Personal board of advisors for AI-assisted code review.

## Advisors

| Name | Focus |
|------|-------|
<!-- Auto-updated by council creator -->

## Usage

Install with [Council CLI](https://github.com/luuuc/council-cli):

` + "```bash" + `
council personas install <this-repo-url>
` + "```" + `

## About

This council was created with [council creator](https://github.com/luuuc/council-cli).
Each markdown file represents an expert persona that can review your code.
`
}

// List returns all custom personas from my-council.
func List() ([]*Persona, error) {
	path, err := MyCouncilPath()
	if err != nil {
		return nil, err
	}

	if !Initialized() {
		return []*Persona{}, nil
	}

	return listPersonasInDir(path, "custom")
}

// listPersonasInDir loads all personas from a directory.
func listPersonasInDir(dir, source string) ([]*Persona, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Persona{}, nil
		}
		return nil, err
	}

	var personas []*Persona
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		// Skip README.md
		if entry.Name() == "README.md" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		p, err := LoadFile(path)
		if err != nil {
			continue // Skip files that can't be parsed
		}
		p.Source = source
		personas = append(personas, p)
	}

	return personas, nil
}

// Load reads a persona by ID from my-council.
func Load(id string) (*Persona, error) {
	path, err := PersonaPath(id)
	if err != nil {
		return nil, err
	}
	return LoadFile(path)
}

// LoadFile reads a persona from a specific file.
func LoadFile(path string) (*Persona, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

// Parse parses persona markdown with frontmatter.
func Parse(data []byte) (*Persona, error) {
	content := string(data)

	// Split frontmatter and body
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("missing frontmatter: file must start with '---'")
	}

	parts := strings.SplitN(content[3:], "---", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid frontmatter: missing closing '---'")
	}

	frontmatter := strings.TrimSpace(parts[0])
	body := strings.TrimSpace(parts[1])

	var p Persona
	if err := yaml.Unmarshal([]byte(frontmatter), &p); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	p.Body = body

	// Set defaults
	if p.Category == "" {
		p.Category = "custom"
	}
	if p.Priority == "" {
		p.Priority = "normal"
	}

	return &p, nil
}

// Save writes a persona to my-council.
func (p *Persona) Save() error {
	if !Initialized() {
		return fmt.Errorf("my-council not initialized: run 'council creator init' first")
	}

	path, err := PersonaPath(p.ID)
	if err != nil {
		return err
	}

	// Generate body if empty
	if p.Body == "" {
		p.Body = p.generateBody()
	}

	// Set defaults
	if p.Category == "" {
		p.Category = "custom"
	}
	if p.Priority == "" {
		p.Priority = "normal"
	}

	// Generate frontmatter
	fm, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal persona: %w", err)
	}

	// Combine frontmatter and body
	content := fmt.Sprintf("---\n%s---\n\n%s", string(fm), p.Body)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write persona file: %w", err)
	}

	return nil
}

func (p *Persona) generateBody() string {
	var buf strings.Builder

	fmt.Fprintf(&buf, "# %s - %s\n\n", p.Name, p.Focus)
	fmt.Fprintf(&buf, "You are channeling %s, known for expertise in %s.\n", p.Name, p.Focus)

	if p.Philosophy != "" {
		fmt.Fprintf(&buf, "\n## Philosophy\n\n%s\n", p.Philosophy)
	}

	if len(p.Principles) > 0 {
		fmt.Fprintf(&buf, "\n## Principles\n\n")
		for _, pr := range p.Principles {
			fmt.Fprintf(&buf, "- %s\n", pr)
		}
	}

	if len(p.RedFlags) > 0 {
		fmt.Fprintf(&buf, "\n## Red Flags\n\nWatch for these patterns:\n")
		for _, rf := range p.RedFlags {
			fmt.Fprintf(&buf, "- %s\n", rf)
		}
	}

	fmt.Fprintf(&buf, "\n## Review Style\n\n")
	fmt.Fprintf(&buf, "When reviewing code, focus on your area of expertise. Be direct and specific.\n")
	fmt.Fprintf(&buf, "Explain your reasoning. Suggest concrete improvements.\n")

	return buf.String()
}

// Path returns the file path for this persona.
func (p *Persona) Path() (string, error) {
	return PersonaPath(p.ID)
}

// Delete removes a persona from my-council and commits.
func Delete(id string) error {
	if !Initialized() {
		return fmt.Errorf("my-council not initialized")
	}

	path, err := PersonaPath(id)
	if err != nil {
		return err
	}

	// Check if exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("persona '%s' not found", id)
	}

	// Load to get name for commit message
	p, err := Load(id)
	if err != nil {
		return err
	}

	// Delete file
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete persona: %w", err)
	}

	// Commit
	myCouncil, err := MyCouncilPath()
	if err != nil {
		return fmt.Errorf("failed to get my-council path: %w", err)
	}
	repo := NewGitRepo(myCouncil)
	if err := repo.Add(id + ".md"); err != nil {
		return err
	}
	return repo.Commit(fmt.Sprintf("Remove persona: %s", p.Name))
}

// Exists checks if a persona exists in my-council.
func Exists(id string) bool {
	path, err := PersonaPath(id)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// SaveAndCommit saves a persona and commits the change.
func (p *Persona) SaveAndCommit(isNew bool) error {
	if err := p.Save(); err != nil {
		return err
	}

	myCouncil, err := MyCouncilPath()
	if err != nil {
		return err
	}

	repo := NewGitRepo(myCouncil)
	action := "Update"
	if isNew {
		action = "Add"
	}
	return repo.AddAndCommit(p.ID+".md", fmt.Sprintf("%s persona: %s", action, p.Name))
}
