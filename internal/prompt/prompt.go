// Package prompt generates AI prompts for council setup based on detected stacks.
package prompt

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/luuuc/council-cli/internal/detect"
)

const setupPromptTemplate = `# Council Setup Request

## Project Analysis

{{if .Languages -}}
**Languages:**
{{range .Languages -}}
- {{.Name}} ({{printf "%.0f" .Percentage}}%)
{{end}}
{{end -}}

{{if .Frameworks -}}
**Frameworks:**
{{range .Frameworks -}}
- {{.Name}}{{if .Version}} {{.Version}}{{end}}
{{end}}
{{end -}}

{{if .Testing -}}
**Testing:**
{{range .Testing -}}
- {{.}}
{{end}}
{{end -}}

{{if .Patterns -}}
**Patterns detected:**
{{range .Patterns -}}
- {{.}}
{{end}}
{{end}}

## Request

Suggest 5-7 expert personas for a code review council. For each expert, provide:

1. A real person known for excellence in a domain relevant to this stack
2. Their specific focus area (1 sentence)
3. A brief philosophy statement in their voice (2-3 paragraphs)
4. Key principles they enforce (bullet points)
5. Red flags they watch for (what makes them intervene)

## Output Format

Return YAML matching this schema:

` + "```yaml" + `
experts:
  - id: short-kebab-case-name
    name: "Full Name"
    focus: "One sentence focus area"
    philosophy: |
      Multi-line philosophy statement written in first person,
      as if the expert is speaking directly to the developer.

      Include their core beliefs about software development
      and what makes code excellent in their domain.
    principles:
      - "First principle they enforce"
      - "Second principle they enforce"
      - "Third principle (aim for 4-6)"
    red_flags:
      - "Pattern that would make them intervene"
      - "Anti-pattern they watch for"
` + "```" + `

## Context

The council will be used to:
- Get expert perspectives during code review (` + "`/council`" + ` command)
- Invoke individual experts for specific questions (` + "`/expert`" + ` commands)
- Enforce consistent quality standards across the project

Choose experts who would naturally complement each other and cover the key aspects of this stack. Prefer well-known figures whose opinions are publicly documented, so their personas can be authentic.
`

// Generate creates a setup prompt from detection results
func Generate(d *detect.Detection) (string, error) {
	tmpl, err := template.New("setup").Parse(setupPromptTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, d); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}
