package review

import (
	"strings"
	"testing"

	"github.com/luuuc/council/internal/expert"
)

func TestSynthesize(t *testing.T) {
	securityExpert := &expert.Expert{
		ID:    "bruce-schneier",
		Name:  "Bruce Schneier",
		Focus: "Application security",
	}
	qualityExpert := &expert.Expert{
		ID:    "kent-beck",
		Name:  "Kent Beck",
		Focus: "Test-driven development",
		Tensions: []expert.Tension{
			{
				Expert:       "jason-fried",
				Topic:        "abstraction",
				Position:     "Extract when the pattern emerges three times",
				Counterpoint: "Don't build for formats nobody asked for",
			},
		},
	}
	scopeExpert := &expert.Expert{
		ID:    "jason-fried",
		Name:  "Jason Fried",
		Focus: "Product simplicity and scope",
	}

	tests := []struct {
		name         string
		verdicts     []ExpertVerdict
		experts      []*expert.Expert
		errors       []string
		wantVerdict  Verdict
		wantBlocking bool
		wantTension  bool
	}{
		{
			name: "all pass",
			verdicts: []ExpertVerdict{
				{Expert: "bruce-schneier", Verdict: VerdictPass, Confidence: 0.9},
				{Expert: "kent-beck", Verdict: VerdictPass, Confidence: 0.85},
			},
			experts:      []*expert.Expert{securityExpert, qualityExpert},
			wantVerdict:  VerdictPass,
			wantBlocking: false,
		},
		{
			name: "one block overrides passes",
			verdicts: []ExpertVerdict{
				{Expert: "bruce-schneier", Verdict: VerdictBlock, Confidence: 0.9, Blocking: true},
				{Expert: "kent-beck", Verdict: VerdictPass, Confidence: 0.85},
			},
			experts:      []*expert.Expert{securityExpert, qualityExpert},
			wantVerdict:  VerdictBlock,
			wantBlocking: true,
		},
		{
			name: "block without blocking flag is not blocking",
			verdicts: []ExpertVerdict{
				{Expert: "bruce-schneier", Verdict: VerdictBlock, Confidence: 0.9, Blocking: false},
				{Expert: "kent-beck", Verdict: VerdictPass, Confidence: 0.85},
			},
			experts:      []*expert.Expert{securityExpert, qualityExpert},
			wantVerdict:  VerdictBlock,
			wantBlocking: false,
		},
		{
			name: "mixed with tension pair",
			verdicts: []ExpertVerdict{
				{Expert: "kent-beck", Verdict: VerdictComment, Confidence: 0.8},
				{Expert: "jason-fried", Verdict: VerdictPass, Confidence: 0.85},
			},
			experts:     []*expert.Expert{qualityExpert, scopeExpert},
			wantVerdict: VerdictComment,
			wantTension: true,
		},
		{
			name: "escalate is highest severity",
			verdicts: []ExpertVerdict{
				{Expert: "bruce-schneier", Verdict: VerdictEscalate, Confidence: 0.5},
				{Expert: "kent-beck", Verdict: VerdictBlock, Confidence: 0.9},
			},
			experts:     []*expert.Expert{securityExpert, qualityExpert},
			wantVerdict: VerdictEscalate,
		},
		{
			name:        "no verdicts with errors",
			verdicts:    nil,
			experts:     nil,
			errors:      []string{"kent-beck: timeout"},
			wantVerdict: VerdictPass, // default
		},
		{
			name: "failed expert excluded from verdict",
			verdicts: []ExpertVerdict{
				{Expert: "kent-beck", Verdict: VerdictPass, Confidence: 0.9},
			},
			experts:     []*expert.Expert{qualityExpert},
			errors:      []string{"bruce-schneier: timeout"},
			wantVerdict: VerdictPass,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Synthesize(tt.verdicts, tt.experts, tt.errors)

			if result.Verdict != tt.wantVerdict {
				t.Errorf("Verdict = %q, want %q", result.Verdict, tt.wantVerdict)
			}
			if result.Blocking != tt.wantBlocking {
				t.Errorf("Blocking = %v, want %v", result.Blocking, tt.wantBlocking)
			}
			if tt.wantTension && result.Tension == "" {
				t.Error("expected Tension to be set, got empty")
			}
			if !tt.wantTension && result.Tension != "" {
				t.Errorf("unexpected Tension: %q", result.Tension)
			}
			if result.Summary == "" {
				t.Error("Summary should not be empty")
			}
		})
	}
}

func TestBuildSummary(t *testing.T) {
	summary := buildSummary(
		[]ExpertVerdict{
			{Expert: "a", Verdict: VerdictPass},
			{Expert: "b", Verdict: VerdictComment},
		},
		[]string{"c: timeout"},
		VerdictComment,
		false,
	)

	if !strings.Contains(summary, "3 experts reviewed") {
		t.Errorf("expected '3 experts reviewed' in summary, got: %s", summary)
	}
	if !strings.Contains(summary, "1 pass") {
		t.Errorf("expected '1 pass' in summary, got: %s", summary)
	}
	if !strings.Contains(summary, "1 failed") {
		t.Errorf("expected '1 failed' in summary, got: %s", summary)
	}
}
