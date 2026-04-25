package review

import (
	"fmt"
	"sort"
	"strings"
)

const (
	// MaxFilesPerReview is the maximum number of files reviewed in a single PR.
	MaxFilesPerReview = 25

	// DefaultTokenBudget is the default per-request input token limit (GitHub Models free tier).
	DefaultTokenBudget = 8000
)

// FileDiff represents a single file's portion of a unified diff.
type FileDiff struct {
	Path       string
	Diff       string
	TokenCount int
	Skipped    bool
	SkipReason string
}

// ChunkResult holds the outcome of splitting a PR diff into per-file chunks.
type ChunkResult struct {
	Files      []FileDiff
	Skipped    []FileDiff
	TotalFiles int
}

// ChunkOptions configures the chunking behavior.
type ChunkOptions struct {
	TokenBudget    int // max input tokens per request (0 = DefaultTokenBudget)
	PromptOverhead int // token estimate for persona prompts added to each call
}

func (o ChunkOptions) effectiveBudget() int {
	budget := o.TokenBudget
	if budget <= 0 {
		budget = DefaultTokenBudget
	}
	budget -= o.PromptOverhead
	if budget < 0 {
		budget = 0
	}
	return budget
}

// EstimateTokens approximates token count from code text using chars/3.
func EstimateTokens(text string) int {
	return (len(text) + 2) / 3
}

// SplitDiff parses a unified diff into per-file chunks, estimates token counts,
// and enforces the file limit and token budget.
func SplitDiff(diff string, opts ChunkOptions) ChunkResult {
	files := parseDiffFiles(diff)
	budget := opts.effectiveBudget()

	for i := range files {
		files[i].TokenCount = EstimateTokens(files[i].Diff)
	}

	// Sort by diff size descending so we review the largest changes first.
	sort.Slice(files, func(i, j int) bool {
		return files[i].TokenCount > files[j].TokenCount
	})

	totalFiles := len(files)

	var reviewable, skipped []FileDiff

	for _, f := range files {
		if f.TokenCount > budget {
			f.Skipped = true
			f.SkipReason = "file too large for per-file review"
			skipped = append(skipped, f)
			continue
		}

		if len(reviewable) >= MaxFilesPerReview {
			f.Skipped = true
			f.SkipReason = fmt.Sprintf("reviewing %d of %d files, largest by diff size", MaxFilesPerReview, totalFiles)
			skipped = append(skipped, f)
			continue
		}

		reviewable = append(reviewable, f)
	}

	return ChunkResult{
		Files:      reviewable,
		Skipped:    skipped,
		TotalFiles: totalFiles,
	}
}

// MergeChunkedResults aggregates per-file SynthesizedResults into one overall result.
func MergeChunkedResults(results []*SynthesizedResult, skipped []FileDiff) *SynthesizedResult {
	merged := &SynthesizedResult{
		Verdict: VerdictPass,
	}

	for _, r := range results {
		if r == nil {
			continue
		}
		if r.Verdict.Severity() > merged.Verdict.Severity() {
			merged.Verdict = r.Verdict
		}
		merged.Perspectives = append(merged.Perspectives, r.Perspectives...)
		merged.Agreements = append(merged.Agreements, r.Agreements...)
		if r.Tension != "" {
			if merged.Tension != "" {
				merged.Tension += "\n\n"
			}
			merged.Tension += r.Tension
		}
		if r.Blocking {
			merged.Blocking = true
		}
		merged.Errors = append(merged.Errors, r.Errors...)
	}

	if len(skipped) > 0 {
		var notes []string
		for _, f := range skipped {
			notes = append(notes, fmt.Sprintf("%s: %s", f.Path, f.SkipReason))
		}
		merged.Summary = fmt.Sprintf("Skipped files:\n%s", strings.Join(notes, "\n"))
	}

	return merged
}

// parseDiffFiles splits a unified diff into per-file FileDiff entries.
func parseDiffFiles(diff string) []FileDiff {
	if diff == "" {
		return nil
	}

	var files []FileDiff
	lines := strings.SplitAfter(diff, "\n")

	var current *FileDiff
	var buf strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git ") {
			if current != nil {
				current.Diff = buf.String()
				files = append(files, *current)
			}
			path := parseDiffPath(line)
			current = &FileDiff{Path: path}
			buf.Reset()
			buf.WriteString(line)
			continue
		}

		if current != nil {
			buf.WriteString(line)
		}
	}

	if current != nil {
		current.Diff = buf.String()
		files = append(files, *current)
	}

	return files
}

// parseDiffPath extracts the file path from a "diff --git a/path b/path" line.
func parseDiffPath(diffLine string) string {
	// Format: "diff --git a/path/to/file b/path/to/file\n"
	diffLine = strings.TrimSpace(diffLine)
	// Split on the last " b/" to handle paths containing " b/" as a substring.
	idx := strings.LastIndex(diffLine, " b/")
	if idx >= 0 {
		return diffLine[idx+3:]
	}
	// Fallback: try to extract from a/ prefix
	parts := strings.SplitN(diffLine, " a/", 2)
	if len(parts) == 2 {
		return strings.SplitN(parts[1], " ", 2)[0]
	}
	return diffLine
}
