package adapter

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/luuuc/council-cli/internal/expert"
)

//go:embed templates/claude/install.md
var claudeInstallTemplate string

//go:embed templates/claude/council-add.md
var claudeCouncilAddTemplate string

//go:embed templates/claude/council-detect.md
var claudeCouncilDetectTemplate string

//go:embed templates/claude/council-remove.md
var claudeCouncilRemoveTemplate string

func init() {
	Register(&Claude{})
}

// Claude is the adapter for Claude Code.
type Claude struct{}

func (c *Claude) Name() string {
	return "claude"
}

func (c *Claude) DisplayName() string {
	return "Claude Code"
}

func (c *Claude) Detect() bool {
	return DirExists(".claude")
}

func (c *Claude) Paths() Paths {
	return Paths{
		Agents:     ".claude/agents",
		Commands:   ".claude/commands",
		Deprecated: []string{},
	}
}

func (c *Claude) Templates() Templates {
	return Templates{
		Install: claudeInstallTemplate,
		Commands: map[string]string{
			"council-add":    claudeCouncilAddTemplate,
			"council-detect": claudeCouncilDetectTemplate,
			"council-remove": claudeCouncilRemoveTemplate,
		},
	}
}

// FormatAgent creates Claude Code agent file content.
// For Claude Code, we use the original expert file content (preserves source format).
func (c *Claude) FormatAgent(e *expert.Expert) string {
	// Read the original expert file and return its content
	data, err := os.ReadFile(e.Path())
	if err != nil {
		// Fallback to regenerating
		return fmt.Sprintf("---\nid: %s\nname: %s\nfocus: %s\n---\n\n%s", e.ID, e.Name, e.Focus, e.Body)
	}
	return string(data)
}

// FormatCommand creates Claude Code command file content.
// Claude Code commands are plain markdown (no frontmatter needed).
func (c *Claude) FormatCommand(name, description, body string) string {
	return body
}

// CouncilCommandTemplate is exported for use by sync when generating the dynamic /council command
func CouncilCommandTemplate() string {
	return `# Code Review Council

Convene the council to review: $ARGUMENTS

## Council Members

{{range .}}
### {{.Name}}
**Focus**: {{.Focus}}
{{end}}

## Instructions

Review the code from each expert's perspective. For each expert:
1. State the expert's name
2. Provide their assessment focused on their domain
3. Note any concerns or suggestions

At the end, synthesize the key points and provide actionable recommendations.
`
}

// agentFilename returns the appropriate filename for an expert based on source
// This is exported for use by sync package
func AgentFilename(e *expert.Expert) string {
	switch {
	case e.Source == "custom":
		return "custom-" + e.ID + ".md"
	case strings.HasPrefix(e.Source, "installed:"):
		return "installed-" + e.ID + ".md"
	default:
		return e.ID + ".md"
	}
}
