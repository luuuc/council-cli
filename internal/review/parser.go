package review

import (
	"encoding/json"
	"regexp"
	"strings"
)

// jsonObjectRe matches a JSON object containing a "verdict" key.
var jsonObjectRe = regexp.MustCompile(`\{[^{}]*"verdict"[^{}]*\}`)

// ParseVerdict extracts a structured ExpertVerdict from raw LLM output.
// It tries multiple extraction strategies in order of specificity,
// falling back to a low-confidence comment verdict if nothing works.
func ParseVerdict(expertID string, raw []byte) ExpertVerdict {
	text := strings.TrimSpace(string(raw))

	if text == "" {
		return fallbackVerdict(expertID, "empty response")
	}

	// Strategy 1: direct JSON unmarshal
	if v, ok := tryUnmarshal(expertID, []byte(text)); ok {
		return v
	}

	// Strategy 2: extract from code fences (```json ... ``` or ``` ... ```)
	if extracted := extractFromCodeFence(text); extracted != "" {
		if v, ok := tryUnmarshal(expertID, []byte(extracted)); ok {
			return v
		}
	}

	// Strategy 3: regex extract JSON object containing "verdict"
	if match := jsonObjectRe.FindString(text); match != "" {
		if v, ok := tryUnmarshal(expertID, []byte(match)); ok {
			return v
		}
	}

	// Strategy 4: fallback
	return fallbackVerdict(expertID, truncate(text, 200))
}

// tryUnmarshal attempts to parse JSON into an ExpertVerdict, validates, and normalizes.
func tryUnmarshal(expertID string, data []byte) (ExpertVerdict, bool) {
	var raw struct {
		Expert     string      `json:"expert"`
		Verdict    Verdict     `json:"verdict"`
		Confidence float64     `json:"confidence"`
		Notes      interface{} `json:"notes"`
		Blocking   bool        `json:"blocking"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return ExpertVerdict{}, false
	}

	// Validate verdict
	if !ValidVerdicts[raw.Verdict] {
		return ExpertVerdict{}, false
	}

	// Normalize confidence to [0, 1]
	if raw.Confidence < 0 {
		raw.Confidence = 0
	}
	if raw.Confidence > 1 {
		raw.Confidence = 1
	}

	// Normalize notes: accept string or []string
	notes := normalizeNotes(raw.Notes)

	return ExpertVerdict{
		Expert:     expertID,
		Verdict:    raw.Verdict,
		Confidence: raw.Confidence,
		Notes:      notes,
		Blocking:   raw.Blocking,
	}, true
}

// normalizeNotes converts various note formats to []string.
func normalizeNotes(v interface{}) []string {
	switch n := v.(type) {
	case []interface{}:
		notes := make([]string, 0, len(n))
		for _, item := range n {
			if s, ok := item.(string); ok && s != "" {
				notes = append(notes, s)
			}
		}
		return notes
	case string:
		if n != "" {
			return []string{n}
		}
		return nil
	default:
		return nil
	}
}

// extractFromCodeFence extracts content from markdown code fences.
func extractFromCodeFence(text string) string {
	// Try ```json first
	if idx := strings.Index(text, "```json"); idx >= 0 {
		content := text[idx+7:]
		if end := strings.Index(content, "```"); end >= 0 {
			return strings.TrimSpace(content[:end])
		}
	}

	// Try plain ```
	if idx := strings.Index(text, "```"); idx >= 0 {
		content := text[idx+3:]
		// Skip optional language tag on same line
		if nl := strings.IndexByte(content, '\n'); nl >= 0 {
			content = content[nl+1:]
		}
		if end := strings.Index(content, "```"); end >= 0 {
			return strings.TrimSpace(content[:end])
		}
	}

	return ""
}

// fallbackVerdict creates a low-confidence comment verdict when parsing fails.
func fallbackVerdict(expertID, rawSnippet string) ExpertVerdict {
	return ExpertVerdict{
		Expert:     expertID,
		Verdict:    VerdictComment,
		Confidence: 0,
		Notes:      []string{"Response could not be parsed: " + rawSnippet},
		Blocking:   false,
		Error:      "unparseable response",
	}
}

// truncate shortens a string to maxLen, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
