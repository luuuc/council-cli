package creator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/luuuc/council-cli/internal/expert"
)

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

// List returns all custom experts from my-council.
func List() ([]*expert.Expert, error) {
	path, err := MyCouncilPath()
	if err != nil {
		return nil, err
	}

	if !Initialized() {
		return []*expert.Expert{}, nil
	}

	return ListExpertsInDir(path, "custom")
}

// ListExpertsInDir loads all experts from a directory.
func ListExpertsInDir(dir, source string) ([]*expert.Expert, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*expert.Expert{}, nil
		}
		return nil, err
	}

	var experts []*expert.Expert
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		// Skip README.md
		if entry.Name() == "README.md" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		e, err := LoadFile(path)
		if err != nil {
			continue // Skip files that can't be parsed
		}
		e.Source = source
		experts = append(experts, e)
	}

	return experts, nil
}

// Load reads an expert by ID from my-council.
func Load(id string) (*expert.Expert, error) {
	path, err := ExpertPath(id)
	if err != nil {
		return nil, err
	}
	return LoadFile(path)
}

// LoadFile reads an expert from a specific file.
func LoadFile(path string) (*expert.Expert, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

// Parse parses expert markdown with frontmatter.
func Parse(data []byte) (*expert.Expert, error) {
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

	e, err := expert.ParseFrontmatter([]byte(frontmatter))
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	e.Body = body
	e.ApplyDefaults()

	return e, nil
}

// Save writes an expert to my-council.
func Save(e *expert.Expert) error {
	if !Initialized() {
		return fmt.Errorf("my-council not initialized: run 'council creator init' first")
	}

	path, err := ExpertPath(e.ID)
	if err != nil {
		return err
	}

	// Generate body if empty
	if e.Body == "" {
		e.Body = GenerateBody(e)
	}

	e.ApplyDefaults()

	// Write using expert's save logic
	return expert.SaveToPath(e, path)
}

// GenerateBody creates default body content for an expert.
func GenerateBody(e *expert.Expert) string {
	var buf strings.Builder

	fmt.Fprintf(&buf, "# %s - %s\n\n", e.Name, e.Focus)
	fmt.Fprintf(&buf, "You are channeling %s, known for expertise in %s.\n", e.Name, e.Focus)

	if e.Philosophy != "" {
		fmt.Fprintf(&buf, "\n## Philosophy\n\n%s\n", e.Philosophy)
	}

	if len(e.Principles) > 0 {
		fmt.Fprintf(&buf, "\n## Principles\n\n")
		for _, pr := range e.Principles {
			fmt.Fprintf(&buf, "- %s\n", pr)
		}
	}

	if len(e.RedFlags) > 0 {
		fmt.Fprintf(&buf, "\n## Red Flags\n\nWatch for these patterns:\n")
		for _, rf := range e.RedFlags {
			fmt.Fprintf(&buf, "- %s\n", rf)
		}
	}

	fmt.Fprintf(&buf, "\n## Review Style\n\n")
	fmt.Fprintf(&buf, "When reviewing code, focus on your area of expertise. Be direct and specific.\n")
	fmt.Fprintf(&buf, "Explain your reasoning. Suggest concrete improvements.\n")

	return buf.String()
}

// ExpertPath returns the full path to an expert file in my-council.
func ExpertPath(id string) (string, error) {
	myCouncil, err := MyCouncilPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(myCouncil, id+".md"), nil
}

// Delete removes an expert from my-council and commits.
func Delete(id string) error {
	if !Initialized() {
		return fmt.Errorf("my-council not initialized")
	}

	path, err := ExpertPath(id)
	if err != nil {
		return err
	}

	// Check if exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("expert '%s' not found - run 'council creator list' to see your personas", id)
	}

	// Load to get name for commit message
	e, err := Load(id)
	if err != nil {
		return err
	}

	// Delete file
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete expert: %w", err)
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
	return repo.Commit(fmt.Sprintf("Remove expert: %s", e.Name))
}

// Exists checks if an expert exists in my-council.
func Exists(id string) bool {
	path, err := ExpertPath(id)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// SaveAndCommit saves an expert and commits the change.
func SaveAndCommit(e *expert.Expert, isNew bool) error {
	if err := Save(e); err != nil {
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
	return repo.AddAndCommit(e.ID+".md", fmt.Sprintf("%s expert: %s", action, e.Name))
}
