package expert

import "testing"

// testBank creates a small suggestion bank for testing
func testBank() SuggestionBank {
	return SuggestionBank{
		"general": {
			{ID: "rob-pike", Name: "Rob Pike", Focus: "Go"},
			{ID: "kent-beck", Name: "Kent Beck", Focus: "Testing"},
			{ID: "dieter-rams", Name: "Dieter Rams", Focus: "Design"},
			{ID: "cal-newport", Name: "Cal Newport", Focus: "Deep Work"},
		},
		"custom": {
			{ID: "luc-perussault-diallo", Name: "Luc Perussault-Diallo", Focus: "Simplicity"},
			{ID: "rob-walling", Name: "Rob Walling", Focus: "SaaS"},
		},
	}
}

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"abc", "abc", 0},
		{"abc", "ab", 1},
		{"ab", "abc", 1},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
		{"rob pike", "rob pik", 1},
	}

	for _, tt := range tests {
		got := levenshtein(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("levenshtein(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestLookupPersona(t *testing.T) {
	bank := testBank()

	tests := []struct {
		input   string
		wantID  string
		wantNil bool
	}{
		// Exact matches
		{"Rob Pike", "rob-pike", false},
		{"rob-pike", "rob-pike", false},
		{"ROB PIKE", "rob-pike", false},
		{"  Rob Pike  ", "rob-pike", false},
		{"Kent Beck", "kent-beck", false},

		// First-name matching (unique first names)
		{"Luc", "luc-perussault-diallo", false},
		{"luc", "luc-perussault-diallo", false},
		{"Dieter", "dieter-rams", false},

		// First-name matching should NOT work for ambiguous names
		// "Rob" could match Rob Pike or Rob Walling
		{"Rob", "", true},

		// Unknown
		{"Unknown Person", "", true},
		{"Brad Pitt", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := LookupPersona(bank, tt.input)
			if tt.wantNil {
				if result != nil {
					t.Errorf("LookupPersona(%q) = %v, want nil", tt.input, result)
				}
			} else {
				if result == nil {
					t.Errorf("LookupPersona(%q) = nil, want ID %q", tt.input, tt.wantID)
				} else if result.ID != tt.wantID {
					t.Errorf("LookupPersona(%q).ID = %q, want %q", tt.input, result.ID, tt.wantID)
				}
			}
		})
	}
}

func TestSuggestSimilar(t *testing.T) {
	bank := testBank()

	tests := []struct {
		input    string
		wantName string // empty means expect nil
	}{
		// Single character typos
		{"kent-bek", "Kent Beck"},
		{"Rob Pik", "Rob Pike"},

		// Case insensitive - exact matches should return nil (use LookupPersona)
		{"ROB PIKE", ""},
		{"rob pike", ""},

		// First-name found by LookupPersona - should return nil
		{"Luc", ""},
		{"luc", ""},
		{"Dieter", ""},
		{"Cal", ""},

		// Prefix matching for short inputs (2-3 chars) - unique prefix
		{"Di", "Dieter Rams"},

		// No close match
		{"xyz", ""},
		{"completely unknown person", ""},

		// Too far (distance > 3)
		{"abcdefgh", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, _ := SuggestSimilar(bank, tt.input)
			if tt.wantName == "" {
				if got != nil {
					t.Errorf("SuggestSimilar(%q) = %q, want nil", tt.input, got.Name)
				}
			} else {
				if got == nil {
					t.Errorf("SuggestSimilar(%q) = nil, want %q", tt.input, tt.wantName)
				} else if got.Name != tt.wantName {
					t.Errorf("SuggestSimilar(%q) = %q, want %q", tt.input, got.Name, tt.wantName)
				}
			}
		})
	}
}

func TestSuggestSimilar_DistanceBoundaries(t *testing.T) {
	bank := testBank()

	tests := []struct {
		input            string
		wantDistance     int
		wantNonNilResult bool
	}{
		// Distance 1 - high confidence
		{"Rob Pik", 1, true},
		{"kent-bek", 1, true},

		// Distance 2 - still prompts
		{"Rob Pi", 2, true},

		// Distance 3 - still matches
		{"Rob P", 3, true},

		// Exact match - returns nil (use LookupPersona instead)
		{"Rob Pike", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, distance := SuggestSimilar(bank, tt.input)
			if tt.wantNonNilResult {
				if got == nil {
					t.Errorf("SuggestSimilar(%q) = nil, want non-nil result", tt.input)
				} else if distance != tt.wantDistance {
					t.Errorf("SuggestSimilar(%q) distance = %d, want %d", tt.input, distance, tt.wantDistance)
				}
			} else {
				if got != nil {
					t.Errorf("SuggestSimilar(%q) = %v, want nil", tt.input, got)
				}
			}
		})
	}
}
