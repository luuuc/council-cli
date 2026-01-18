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
