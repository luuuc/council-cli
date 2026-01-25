// Package export formats expert councils for use outside the council-cli ecosystem.
package export

import (
	"strings"

	"github.com/luuuc/council-cli/internal/expert"
)

// FormatMarkdown generates portable markdown for use in any AI context
func FormatMarkdown(experts []*expert.Expert) string {
	var b strings.Builder

	b.WriteString("# Expert Council\n\n")
	b.WriteString("Use these expert perspectives when reviewing my work.\n\n")

	for i, e := range experts {
		b.WriteString("## ")
		b.WriteString(e.Name)
		b.WriteString("\n")

		b.WriteString("**Focus**: ")
		b.WriteString(e.Focus)
		b.WriteString("\n\n")

		if e.Philosophy != "" {
			b.WriteString(strings.TrimSpace(e.Philosophy))
			b.WriteString("\n\n")
		}

		if len(e.Principles) > 0 {
			b.WriteString("**Principles**:\n")
			for _, p := range e.Principles {
				b.WriteString("- ")
				b.WriteString(p)
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}

		if len(e.RedFlags) > 0 {
			b.WriteString("**Watch for**:\n")
			for _, r := range e.RedFlags {
				b.WriteString("- ")
				b.WriteString(r)
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}

		// Add separator between experts (but not after the last one)
		if i < len(experts)-1 {
			b.WriteString("---\n\n")
		}
	}

	return b.String()
}
