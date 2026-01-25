package cmd

import "testing"

func TestFindYAMLStart(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "starts with ---",
			input:    "---\nid: test\n---",
			expected: 0,
		},
		{
			name:     "--- after text",
			input:    "Some preamble\n---\nid: test\n---",
			expected: 14,
		},
		{
			name:     "--- after code block",
			input:    "```yaml\n---\nid: test\n---\n```",
			expected: 8,
		},
		{
			name:     "no ---",
			input:    "just some text",
			expected: -1,
		},
		{
			name:     "empty string",
			input:    "",
			expected: -1,
		},
		{
			name:     "--- in middle of line (not valid)",
			input:    "text---more",
			expected: -1,
		},
		{
			name:     "multiple ---",
			input:    "---\nfirst\n---\nsecond\n---",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findYAMLStart(tt.input)
			if result != tt.expected {
				t.Errorf("findYAMLStart(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "no truncation needed",
			input:    "short",
			maxLen:   10,
			expected: "short",
		},
		{
			name:     "exact length",
			input:    "exact",
			maxLen:   5,
			expected: "exact",
		},
		{
			name:     "needs truncation",
			input:    "this is a long string",
			maxLen:   10,
			expected: "this is...",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "very short maxLen",
			input:    "hello",
			maxLen:   4,
			expected: "h...",
		},
		{
			name:     "unicode string",
			input:    "héllo wörld",
			maxLen:   8,
			expected: "héll...", // truncate counts bytes, not runes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, expected %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected []string
	}{
		{
			name:     "no wrapping needed",
			input:    "short text",
			width:    20,
			expected: []string{"short text"},
		},
		{
			name:     "simple wrap",
			input:    "this is a longer piece of text",
			width:    15,
			expected: []string{"this is a", "longer piece of", "text"},
		},
		{
			name:     "empty string",
			input:    "",
			width:    10,
			expected: nil,
		},
		{
			name:     "single word longer than width",
			input:    "superlongword",
			width:    5,
			expected: []string{"superlongword"},
		},
		{
			name:     "multiple spaces collapsed",
			input:    "word1   word2    word3",
			width:    20,
			expected: []string{"word1 word2 word3"},
		},
		{
			name:     "newlines treated as spaces",
			input:    "line1\nline2\nline3",
			width:    20,
			expected: []string{"line1 line2 line3"},
		},
		{
			name:     "tabs treated as spaces",
			input:    "col1\tcol2\tcol3",
			width:    20,
			expected: []string{"col1 col2 col3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapText(tt.input, tt.width)
			if len(result) != len(tt.expected) {
				t.Errorf("wrapText(%q, %d) returned %d lines, expected %d", tt.input, tt.width, len(result), len(tt.expected))
				t.Errorf("got: %v", result)
				t.Errorf("expected: %v", tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("wrapText(%q, %d)[%d] = %q, expected %q", tt.input, tt.width, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

