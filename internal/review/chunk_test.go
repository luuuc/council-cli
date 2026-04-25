package review

import (
	"fmt"
	"strings"
	"testing"
)

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"abc", 1},
		{"abcdef", 2},
		{strings.Repeat("x", 300), 100},
		{strings.Repeat("x", 3000), 1000},
	}
	for _, tt := range tests {
		got := EstimateTokens(tt.input)
		if got != tt.want {
			t.Errorf("EstimateTokens(%d chars) = %d, want %d", len(tt.input), got, tt.want)
		}
	}
}

func TestParseDiffFiles(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main
+import "fmt"
 func main() {}
diff --git a/util.go b/util.go
--- a/util.go
+++ b/util.go
@@ -1,2 +1,3 @@
 package main
+func helper() {}
`

	files := parseDiffFiles(diff)
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0].Path != "main.go" {
		t.Errorf("first file path = %q, want main.go", files[0].Path)
	}
	if files[1].Path != "util.go" {
		t.Errorf("second file path = %q, want util.go", files[1].Path)
	}
	if !strings.Contains(files[0].Diff, "import \"fmt\"") {
		t.Error("first file diff should contain the added import")
	}
	if !strings.Contains(files[1].Diff, "func helper()") {
		t.Error("second file diff should contain the added function")
	}
}

func TestParseDiffFilesEmpty(t *testing.T) {
	files := parseDiffFiles("")
	if len(files) != 0 {
		t.Errorf("expected 0 files for empty diff, got %d", len(files))
	}
}

func TestParseDiffPath(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"diff --git a/src/main.go b/src/main.go", "src/main.go"},
		{"diff --git a/README.md b/README.md", "README.md"},
		{"diff --git a/a/b/c.txt b/a/b/c.txt", "a/b/c.txt"},
	}
	for _, tt := range tests {
		got := parseDiffPath(tt.line)
		if got != tt.want {
			t.Errorf("parseDiffPath(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestSplitDiffSmallFiles(t *testing.T) {
	diff := makeDiff(10, 100)
	result := SplitDiff(diff, ChunkOptions{TokenBudget: 8000})

	if len(result.Files) != 10 {
		t.Errorf("expected 10 reviewable files, got %d", len(result.Files))
	}
	if len(result.Skipped) != 0 {
		t.Errorf("expected 0 skipped files, got %d", len(result.Skipped))
	}
	if result.TotalFiles != 10 {
		t.Errorf("expected TotalFiles=10, got %d", result.TotalFiles)
	}
}

func TestSplitDiffOversizedFile(t *testing.T) {
	// One file with 30000 chars (~10000 tokens), rest are small
	diff := makeDiffMixed(5, 100, 1, 30000)
	result := SplitDiff(diff, ChunkOptions{TokenBudget: 8000})

	if len(result.Skipped) != 1 {
		t.Fatalf("expected 1 skipped file, got %d", len(result.Skipped))
	}
	if result.Skipped[0].SkipReason != "file too large for per-file review" {
		t.Errorf("unexpected skip reason: %q", result.Skipped[0].SkipReason)
	}
	if len(result.Files) != 5 {
		t.Errorf("expected 5 reviewable files, got %d", len(result.Files))
	}
}

func TestSplitDiffFileLimit(t *testing.T) {
	diff := makeDiff(30, 100)
	result := SplitDiff(diff, ChunkOptions{TokenBudget: 8000})

	if len(result.Files) != MaxFilesPerReview {
		t.Errorf("expected %d reviewable files, got %d", MaxFilesPerReview, len(result.Files))
	}
	if len(result.Skipped) != 5 {
		t.Errorf("expected 5 skipped files, got %d", len(result.Skipped))
	}
	if result.TotalFiles != 30 {
		t.Errorf("expected TotalFiles=30, got %d", result.TotalFiles)
	}
	for _, s := range result.Skipped {
		if !strings.Contains(s.SkipReason, "reviewing 25 of 30 files") {
			t.Errorf("expected file limit skip reason, got: %q", s.SkipReason)
		}
	}
}

func TestSplitDiffPromptOverhead(t *testing.T) {
	// File with ~2500 tokens fits in 8000 budget but not if 6000 tokens of prompt overhead
	diff := makeDiff(1, 7500) // ~2500 tokens
	result := SplitDiff(diff, ChunkOptions{TokenBudget: 8000, PromptOverhead: 6000})

	if len(result.Skipped) != 1 {
		t.Errorf("expected 1 skipped (budget=2000 after overhead), got %d skipped", len(result.Skipped))
	}
}

func TestSplitDiffNegativeBudget(t *testing.T) {
	diff := makeDiff(3, 100)
	result := SplitDiff(diff, ChunkOptions{TokenBudget: 100, PromptOverhead: 200})

	// Budget is clamped to 0, so all files should be skipped
	if len(result.Files) != 0 {
		t.Errorf("expected 0 reviewable files with negative effective budget, got %d", len(result.Files))
	}
	if len(result.Skipped) != 3 {
		t.Errorf("expected 3 skipped files, got %d", len(result.Skipped))
	}
}

func TestParseDiffPathEdgeCases(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"diff --git a/src/main.go b/src/main.go", "src/main.go"},
		{"diff --git a/README.md b/README.md", "README.md"},
		{"diff --git a/deep/nested/path/file.go b/deep/nested/path/file.go", "deep/nested/path/file.go"},
		{"diff --git a/b/file.go b/b/file.go", "b/file.go"},
	}
	for _, tt := range tests {
		got := parseDiffPath(tt.line)
		if got != tt.want {
			t.Errorf("parseDiffPath(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestSplitDiffSortsBySize(t *testing.T) {
	diff := `diff --git a/small.go b/small.go
--- a/small.go
+++ b/small.go
@@ -1 +1,2 @@
+x
diff --git a/large.go b/large.go
--- a/large.go
+++ b/large.go
@@ -1 +1,50 @@
` + strings.Repeat("+line\n", 49)

	result := SplitDiff(diff, ChunkOptions{TokenBudget: 8000})
	if len(result.Files) < 2 {
		t.Fatal("expected at least 2 files")
	}
	if result.Files[0].Path != "large.go" {
		t.Errorf("expected largest file first, got %q", result.Files[0].Path)
	}
}

func TestMergeChunkedResults(t *testing.T) {
	r1 := &SynthesizedResult{
		Verdict:      VerdictPass,
		Perspectives: []ExpertVerdict{{Expert: "a", Verdict: VerdictPass}},
		Agreements:   []string{"Clean code"},
	}
	r2 := &SynthesizedResult{
		Verdict:      VerdictComment,
		Perspectives: []ExpertVerdict{{Expert: "b", Verdict: VerdictComment}},
		Tension:      "A disagrees with B",
	}
	r3 := &SynthesizedResult{
		Verdict:  VerdictBlock,
		Blocking: true,
		Errors:   []string{"timeout on file3"},
	}

	skipped := []FileDiff{
		{Path: "big.go", SkipReason: "file too large for per-file review"},
	}

	merged := MergeChunkedResults([]*SynthesizedResult{r1, r2, r3}, skipped)

	if merged.Verdict != VerdictBlock {
		t.Errorf("merged verdict = %s, want block", merged.Verdict)
	}
	if !merged.Blocking {
		t.Error("merged should be blocking")
	}
	if len(merged.Perspectives) != 2 {
		t.Errorf("expected 2 perspectives, got %d", len(merged.Perspectives))
	}
	if len(merged.Agreements) != 1 {
		t.Errorf("expected 1 agreement, got %d", len(merged.Agreements))
	}
	if merged.Tension != "A disagrees with B" {
		t.Errorf("unexpected tension: %q", merged.Tension)
	}
	if len(merged.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(merged.Errors))
	}
	if !strings.Contains(merged.Summary, "big.go") {
		t.Error("summary should mention skipped file")
	}
}

func TestMergeChunkedResultsNilResults(t *testing.T) {
	merged := MergeChunkedResults([]*SynthesizedResult{nil, nil}, nil)
	if merged.Verdict != VerdictPass {
		t.Errorf("merged nil results should default to pass, got %s", merged.Verdict)
	}
}

// --- helpers ---

func makeDiff(numFiles, charsPerFile int) string {
	var b strings.Builder
	for i := 0; i < numFiles; i++ {
		fmt.Fprintf(&b, "diff --git a/file%d.go b/file%d.go\n", i, i)
		fmt.Fprintf(&b, "--- a/file%d.go\n+++ b/file%d.go\n", i, i)
		b.WriteString("@@ -1 +1,2 @@\n")
		b.WriteString("+" + strings.Repeat("x", charsPerFile) + "\n")
	}
	return b.String()
}

func makeDiffMixed(numSmall, smallChars, numLarge, largeChars int) string {
	var b strings.Builder
	for i := 0; i < numSmall; i++ {
		fmt.Fprintf(&b, "diff --git a/small%d.go b/small%d.go\n", i, i)
		fmt.Fprintf(&b, "--- a/small%d.go\n+++ b/small%d.go\n", i, i)
		b.WriteString("@@ -1 +1,2 @@\n")
		b.WriteString("+" + strings.Repeat("x", smallChars) + "\n")
	}
	for i := 0; i < numLarge; i++ {
		fmt.Fprintf(&b, "diff --git a/large%d.go b/large%d.go\n", i, i)
		fmt.Fprintf(&b, "--- a/large%d.go\n+++ b/large%d.go\n", i, i)
		b.WriteString("@@ -1 +1,2 @@\n")
		b.WriteString("+" + strings.Repeat("x", largeChars) + "\n")
	}
	return b.String()
}
