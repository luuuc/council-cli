package cmd

import "testing"

func TestSuggestionsSchema(t *testing.T) {
	// suggestionBank is loaded in init() from suggestions.yaml

	if len(suggestionBank) == 0 {
		t.Fatal("suggestionBank is empty - suggestions.yaml failed to load")
	}

	seenIDs := make(map[string]string) // id -> category for duplicate detection
	totalExperts := 0

	for category, experts := range suggestionBank {
		if len(experts) == 0 {
			t.Errorf("category %q has no experts", category)
			continue
		}

		for i, e := range experts {
			totalExperts++
			prefix := func(msg string) string {
				return category + "[" + string(rune('0'+i)) + "] " + e.Name + ": " + msg
			}

			// Required fields
			if e.ID == "" {
				t.Error(prefix("missing id"))
			}
			if e.Name == "" {
				t.Error(prefix("missing name"))
			}
			if e.Focus == "" {
				t.Error(prefix("missing focus"))
			}
			if e.Philosophy == "" {
				t.Error(prefix("missing philosophy"))
			}
			if len(e.Principles) == 0 {
				t.Error(prefix("missing principles"))
			}
			if len(e.RedFlags) == 0 {
				t.Error(prefix("missing red_flags"))
			}

			// Quality checks
			if len(e.Principles) < 3 {
				t.Errorf(prefix("too few principles (%d, want at least 3)"), len(e.Principles))
			}
			if len(e.RedFlags) < 2 {
				t.Errorf(prefix("too few red_flags (%d, want at least 2)"), len(e.RedFlags))
			}

			// ID format: must be kebab-case (lowercase, hyphens only)
			// Note: IDs can differ from ToID(name) for aliases (tenderlove),
			// disambiguation (jose-valim vs jose-valim-phoenix), or accented names
			for _, c := range e.ID {
				if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
					t.Errorf(prefix("id %q contains invalid character %q"), e.ID, string(c))
					break
				}
			}

			// Duplicate ID check
			if prevCategory, exists := seenIDs[e.ID]; exists {
				t.Errorf(prefix("duplicate id %q (also in %s)"), e.ID, prevCategory)
			}
			seenIDs[e.ID] = category
		}
	}

	t.Logf("Validated %d experts across %d categories", totalExperts, len(suggestionBank))
}

func TestSuggestionsTriggers(t *testing.T) {
	// Check that general experts have either Core=true or Triggers defined
	generalExperts := suggestionBank["general"]
	if len(generalExperts) == 0 {
		t.Skip("no general category found")
	}

	for _, e := range generalExperts {
		if !e.Core && len(e.Triggers) == 0 {
			t.Errorf("general expert %q has neither core=true nor triggers defined", e.Name)
		}
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

func TestSuggestSimilar(t *testing.T) {
	tests := []struct {
		input    string
		wantName string // empty means expect nil
	}{
		// Single character typos
		{"Luc Perussault-Diall", "Luc Perussault-Diallo"},
		{"kent-bek", "Kent Beck"},
		{"Rob Pik", "Rob Pike"},

		// Case insensitive - exact matches should return nil (use LookupPersona)
		{"ROB PIKE", ""},
		{"rob pike", ""},

		// First-name found by LookupPersona - should return nil
		{"Luc", ""},    // LookupPersona finds this now
		{"luc", ""},    // LookupPersona finds this now
		{"Dieter", ""},  // LookupPersona finds this now
		{"Cal", ""},     // LookupPersona finds Cal Newport (exact first-name match)

		// Prefix matching for short inputs (2-3 chars) - unique prefix
		{"Di", "Dieter Rams"}, // "Di" prefix matches only Dieter Rams

		// No close match
		{"xyz", ""},
		{"completely unknown person", ""},

		// Too far (distance > 3)
		{"abcdefgh", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, _ := SuggestSimilar(tt.input)
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

func TestLookupPersona(t *testing.T) {
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
		// (e.g., "Rob" could match multiple people - Rob Pike, Rob Walling)
		// This returns nil because there are multiple matches
		{"Rob", "", true},

		// Unknown
		{"Unknown Person", "", true},
		{"Brad Pitt", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := LookupPersona(tt.input)
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
