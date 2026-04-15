package review

import (
	"bytes"
	"text/template"

	"github.com/luuuc/council-cli/internal/expert"
)

var promptTemplate = template.Must(template.New("review-prompt").Parse(`You are {{.Expert.Name}}, reviewing code as part of a council review.

## Your Persona

{{.Expert.Body}}

## Submission

` + "```" + `
{{.Submission.Content}}
` + "```" + `
{{if .Submission.Context}}
## Context

{{.Submission.Context}}
{{end}}
## Response Format

You MUST respond with ONLY a JSON object matching this exact schema. No markdown, no code fences, no explanation before or after.

{"expert":"{{.Expert.ID}}","verdict":"<pass|comment|block|escalate>","confidence":<0.0-1.0>,"notes":["<observation 1>","<observation 2>"],"blocking":false}

Field definitions:
- verdict: "pass" (no issues), "comment" (suggestions worth considering), "block" (must fix before shipping), "escalate" (beyond your expertise to judge)
- confidence: how confident you are in your assessment, from 0.0 to 1.0
- notes: specific observations from your area of expertise — be direct and concrete
- blocking: true only if this is a blocking issue that must be resolved

Respond with ONLY the JSON object. Nothing else.`))

type promptData struct {
	Expert     *expert.Expert
	Submission Submission
}

// BuildPrompt constructs the review prompt for an expert and submission.
func BuildPrompt(e *expert.Expert, sub Submission) string {
	var buf bytes.Buffer
	if err := promptTemplate.Execute(&buf, promptData{Expert: e, Submission: sub}); err != nil {
		// Fallback to a minimal prompt if template fails
		return "Review this code as " + e.Name + ":\n\n" + sub.Content
	}
	return buf.String()
}
