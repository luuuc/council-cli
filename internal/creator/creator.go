package creator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/luuuc/council-cli/internal/expert"
)

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
