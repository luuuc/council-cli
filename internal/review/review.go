// Package review implements the council review engine: blind parallel expert
// reviews with structured verdict parsing and tension-aware synthesis.
package review

// Verdict represents the possible review outcomes.
type Verdict string

const (
	VerdictPass     Verdict = "pass"
	VerdictComment  Verdict = "comment"
	VerdictBlock    Verdict = "block"
	VerdictEscalate Verdict = "escalate"
)

// ValidVerdicts is the set of recognized verdict values.
var ValidVerdicts = map[Verdict]bool{
	VerdictPass:     true,
	VerdictComment:  true,
	VerdictBlock:    true,
	VerdictEscalate: true,
}

// Severity returns the numeric severity for ordering verdicts.
// Higher is more severe.
func (v Verdict) Severity() int {
	switch v {
	case VerdictPass:
		return 0
	case VerdictComment:
		return 1
	case VerdictBlock:
		return 2
	case VerdictEscalate:
		return 3
	default:
		return 1 // unknown defaults to comment-level
	}
}

// ExpertVerdict is the structured output from a single expert review.
type ExpertVerdict struct {
	Expert     string   `json:"expert"`
	Verdict    Verdict  `json:"verdict"`
	Confidence float64  `json:"confidence"`
	Notes      []string `json:"notes"`
	Blocking   bool     `json:"blocking"`
	Error      string   `json:"error,omitempty"`
}

// Submission is the material being reviewed.
type Submission struct {
	Content   string // The diff, file content, or text to review
	Context   string // Optional context (e.g., PR title)
	RawPrompt string // When set, backends use this as the prompt directly (bypasses BuildPrompt and ParseVerdict)
}

// SynthesizedResult is the aggregated output from all expert reviews.
type SynthesizedResult struct {
	Verdict      Verdict         `json:"verdict"`
	Blocking     bool            `json:"blocking"`
	Perspectives []ExpertVerdict `json:"perspectives"`
	Agreements   []string        `json:"agreements"`
	Tension      string          `json:"tension"`
	Summary      string          `json:"summary"`
	Errors       []string        `json:"errors,omitempty"`
}

// ReviewOptions controls review execution.
type ReviewOptions struct {
	Concurrency int
	Timeout     int // per-expert timeout in seconds
}

// DefaultConcurrency is the default number of parallel expert reviews.
// Informed by the assumption that 4 concurrent CLI subprocesses are viable
// on a 16GB developer machine. Adjustable via --concurrency flag.
const DefaultConcurrency = 4
