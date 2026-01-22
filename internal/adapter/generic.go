package adapter

import (
	"fmt"
	"strings"

	"github.com/luuuc/council-cli/internal/expert"
)

func init() {
	Register(&Generic{})
}

// Generic is the fallback adapter for projects without a specific AI tool.
// It generates an AGENTS.md file in the project root.
type Generic struct{}

func (g *Generic) Name() string {
	return "generic"
}

func (g *Generic) DisplayName() string {
	return "Generic (AGENTS.md)"
}

// Detect always returns true - generic is the fallback.
// However, it's excluded from automatic detection in Detect()
// and must be explicitly selected.
func (g *Generic) Detect() bool {
	return true
}

func (g *Generic) Paths() Paths {
	return Paths{
		Agents:     ".", // AGENTS.md in project root
		Commands:   ".", // No separate commands
		Deprecated: []string{},
	}
}

func (g *Generic) Templates() Templates {
	return Templates{
		Install:  genericInstallTemplate,
		Commands: map[string]string{}, // No commands for generic
	}
}

// FormatAgent creates a simple markdown section for an expert.
// For generic, this is used as part of AGENTS.md generation.
func (g *Generic) FormatAgent(e *expert.Expert) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("### %s%s", e.Name, e.SourceMarker()))
	parts = append(parts, fmt.Sprintf("- **ID**: %s", e.ID))
	parts = append(parts, fmt.Sprintf("- **Focus**: %s", e.Focus))
	parts = append(parts, "")

	if e.Philosophy != "" {
		parts = append(parts, strings.TrimSpace(e.Philosophy))
		parts = append(parts, "")
	}

	if len(e.Principles) > 0 {
		parts = append(parts, "**Principles:**")
		for _, p := range e.Principles {
			parts = append(parts, fmt.Sprintf("- %s", p))
		}
		parts = append(parts, "")
	}

	return strings.Join(parts, "\n")
}

// FormatCommand returns empty for generic - no commands supported.
func (g *Generic) FormatCommand(name, description, body string) string {
	return ""
}

// GenerateAgentsMd creates the complete AGENTS.md file content.
// This is a special method for the generic adapter since it combines
// all experts into a single file rather than separate files.
func (g *Generic) GenerateAgentsMd(experts []*expert.Expert) string {
	var parts []string

	parts = append(parts, "# AGENTS.md - Expert Council")
	parts = append(parts, "")
	parts = append(parts, "This file defines expert personas for AI coding assistants.")
	parts = append(parts, "")
	parts = append(parts, "## Council Members")
	parts = append(parts, "")

	for _, e := range experts {
		parts = append(parts, g.FormatAgent(e))
	}

	return strings.Join(parts, "\n")
}

var genericInstallTemplate = `# Install Council

Set up the council for your project.

## Quick Start

1. Initialize the council:
` + "```bash" + `
council init
` + "```" + `

2. Add experts to your council:
` + "```bash" + `
council add "Kent Beck"
council add "Sandi Metz"
` + "```" + `

3. Sync to generate AGENTS.md:
` + "```bash" + `
council sync
` + "```" + `

The AGENTS.md file will be created in your project root.
AI tools that support the AGENTS.md convention will use these expert personas.
`
