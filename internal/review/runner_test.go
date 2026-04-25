package review

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/luuuc/council/internal/expert"
)

// MockBackend returns canned verdicts for testing.
type MockBackend struct {
	Results           map[string]ExpertVerdict
	Errors            map[string]error
	CollectiveResult  *SynthesizedResult
	CollectiveErr     error
	Delay             time.Duration
	calls             atomic.Int32
	collectiveCalls   atomic.Int32
}

func (m *MockBackend) Review(ctx context.Context, e *expert.Expert, sub Submission) (ExpertVerdict, error) {
	m.calls.Add(1)

	if m.Delay > 0 {
		select {
		case <-time.After(m.Delay):
		case <-ctx.Done():
			return ExpertVerdict{}, ctx.Err()
		}
	}

	if err, ok := m.Errors[e.ID]; ok {
		return ExpertVerdict{}, err
	}

	if v, ok := m.Results[e.ID]; ok {
		return v, nil
	}

	return ExpertVerdict{
		Expert:     e.ID,
		Verdict:    VerdictPass,
		Confidence: 0.9,
	}, nil
}

func (m *MockBackend) ReviewCollective(ctx context.Context, experts []*expert.Expert, sub Submission) (*SynthesizedResult, error) {
	m.collectiveCalls.Add(1)

	if m.Delay > 0 {
		select {
		case <-time.After(m.Delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.CollectiveErr != nil {
		return nil, m.CollectiveErr
	}

	if m.CollectiveResult != nil {
		return m.CollectiveResult, nil
	}

	perspectives := make([]ExpertVerdict, len(experts))
	for i, e := range experts {
		perspectives[i] = ExpertVerdict{
			Expert: e.ID, Verdict: VerdictPass, Confidence: 0.9,
		}
	}
	return &SynthesizedResult{
		Verdict:      VerdictPass,
		Perspectives: perspectives,
		Summary:      "All good.",
	}, nil
}

func TestRunnerCollectiveHappyPath(t *testing.T) {
	backend := &MockBackend{
		CollectiveResult: &SynthesizedResult{
			Verdict: VerdictComment,
			Perspectives: []ExpertVerdict{
				{Expert: "kent-beck", Verdict: VerdictComment, Confidence: 0.8, Notes: []string{"Add test"}},
				{Expert: "bruce-schneier", Verdict: VerdictPass, Confidence: 0.95},
			},
			Agreements: []string{"Code structure is clean"},
			Tension:    "Kent wants more tests, Bruce is satisfied with security",
			Summary:    "Ship with comments.",
		},
	}

	runner := &Runner{
		Backend: backend,
		Options: ReviewOptions{Timeout: 10},
	}

	inputs := []ExpertInput{
		{Expert: &expert.Expert{ID: "kent-beck", Name: "Kent Beck", Focus: "TDD"}, Blocking: false},
		{Expert: &expert.Expert{ID: "bruce-schneier", Name: "Bruce Schneier", Focus: "Security"}, Blocking: true},
	}

	result := runner.Run(context.Background(), inputs, Submission{Content: "test diff"})

	if backend.collectiveCalls.Load() != 1 {
		t.Errorf("expected 1 ReviewCollective call, got %d", backend.collectiveCalls.Load())
	}
	if backend.calls.Load() != 0 {
		t.Errorf("expected 0 Review calls, got %d", backend.calls.Load())
	}
	if len(result.Perspectives) != 2 {
		t.Fatalf("expected 2 perspectives, got %d", len(result.Perspectives))
	}
	if result.Verdict != VerdictComment {
		t.Errorf("expected comment verdict, got %s", result.Verdict)
	}
	if result.Tension == "" {
		t.Error("expected tension to be preserved from LLM response")
	}
}

func TestRunnerCollectiveBlockingFromPackConfig(t *testing.T) {
	backend := &MockBackend{
		CollectiveResult: &SynthesizedResult{
			Verdict: VerdictBlock,
			Perspectives: []ExpertVerdict{
				{Expert: "kent-beck", Verdict: VerdictPass, Confidence: 0.9, Blocking: true},
				{Expert: "bruce-schneier", Verdict: VerdictBlock, Confidence: 0.9, Blocking: false},
			},
			Summary: "Block.",
		},
	}

	runner := &Runner{
		Backend: backend,
		Options: ReviewOptions{Timeout: 10},
	}

	inputs := []ExpertInput{
		{Expert: &expert.Expert{ID: "kent-beck", Name: "Kent Beck", Focus: "TDD"}, Blocking: false},
		{Expert: &expert.Expert{ID: "bruce-schneier", Name: "Bruce Schneier", Focus: "Security"}, Blocking: true},
	}

	result := runner.Run(context.Background(), inputs, Submission{Content: "test diff"})

	// kent-beck: LLM said blocking=true but pack says false
	if result.Perspectives[0].Blocking {
		t.Error("kent-beck blocking should be overridden to false from pack config")
	}
	// bruce-schneier: LLM said blocking=false but pack says true
	if !result.Perspectives[1].Blocking {
		t.Error("bruce-schneier blocking should be overridden to true from pack config")
	}
	// With bruce-schneier blocking + VerdictBlock, result should be blocking
	if !result.Blocking {
		t.Error("expected overall result to be blocking")
	}
}

func TestRunnerCollectiveHierarchyOverride(t *testing.T) {
	backend := &MockBackend{
		CollectiveResult: &SynthesizedResult{
			Verdict: VerdictPass, // LLM incorrectly says pass
			Perspectives: []ExpertVerdict{
				{Expert: "kent-beck", Verdict: VerdictPass, Confidence: 0.9},
				{Expert: "bruce-schneier", Verdict: VerdictBlock, Confidence: 0.9},
			},
			Summary: "Ship it.",
		},
	}

	runner := &Runner{
		Backend: backend,
		Options: ReviewOptions{Timeout: 10},
	}

	inputs := []ExpertInput{
		{Expert: &expert.Expert{ID: "kent-beck", Name: "Kent Beck", Focus: "TDD"}, Blocking: false},
		{Expert: &expert.Expert{ID: "bruce-schneier", Name: "Bruce Schneier", Focus: "Security"}, Blocking: true},
	}

	result := runner.Run(context.Background(), inputs, Submission{Content: "test diff"})

	if result.Verdict != VerdictBlock {
		t.Errorf("hierarchy should override LLM verdict to block, got %s", result.Verdict)
	}
}

func TestRunnerSingleExpertUsesPerExpertPath(t *testing.T) {
	backend := &MockBackend{
		Results: map[string]ExpertVerdict{
			"kent-beck": {Expert: "kent-beck", Verdict: VerdictPass, Confidence: 0.9},
		},
	}

	runner := &Runner{
		Backend: backend,
		Options: ReviewOptions{Concurrency: 2, Timeout: 10},
	}

	inputs := []ExpertInput{
		{Expert: &expert.Expert{ID: "kent-beck", Name: "Kent Beck", Focus: "TDD"}},
	}

	result := runner.Run(context.Background(), inputs, Submission{Content: "test diff"})

	if backend.calls.Load() != 1 {
		t.Errorf("expected 1 Review call for single expert, got %d", backend.calls.Load())
	}
	if backend.collectiveCalls.Load() != 0 {
		t.Errorf("expected 0 ReviewCollective calls for single expert, got %d", backend.collectiveCalls.Load())
	}
	if result.Verdict != VerdictPass {
		t.Errorf("expected pass, got %s", result.Verdict)
	}
}

func TestRunnerFallbackOnLargePrompt(t *testing.T) {
	backend := &MockBackend{
		Results: map[string]ExpertVerdict{
			"expert-a": {Expert: "expert-a", Verdict: VerdictPass, Confidence: 0.9},
			"expert-b": {Expert: "expert-b", Verdict: VerdictPass, Confidence: 0.9},
		},
	}

	runner := &Runner{
		Backend: backend,
		Options: ReviewOptions{Concurrency: 2, Timeout: 10},
	}

	// Create a submission large enough to exceed CollectiveThreshold
	largeContent := strings.Repeat("x", CollectiveThreshold)

	inputs := []ExpertInput{
		{Expert: &expert.Expert{ID: "expert-a", Name: "Expert A", Focus: "Testing", Body: "Expert A body."}},
		{Expert: &expert.Expert{ID: "expert-b", Name: "Expert B", Focus: "Testing", Body: "Expert B body."}},
	}

	result := runner.Run(context.Background(), inputs, Submission{Content: largeContent})

	if backend.calls.Load() != 2 {
		t.Errorf("expected 2 per-expert Review calls for fallback, got %d", backend.calls.Load())
	}
	if backend.collectiveCalls.Load() != 0 {
		t.Errorf("expected 0 ReviewCollective calls for fallback, got %d", backend.collectiveCalls.Load())
	}
	if len(result.Perspectives) != 2 {
		t.Errorf("expected 2 perspectives, got %d", len(result.Perspectives))
	}
}

func TestRunnerCollectiveContextCancellation(t *testing.T) {
	backend := &MockBackend{
		Delay: 5 * time.Second,
	}

	runner := &Runner{
		Backend: backend,
		Options: ReviewOptions{Timeout: 1},
	}

	inputs := []ExpertInput{
		{Expert: &expert.Expert{ID: "kent-beck", Name: "Kent Beck", Focus: "TDD"}},
		{Expert: &expert.Expert{ID: "bruce-schneier", Name: "Bruce Schneier", Focus: "Security"}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result := runner.Run(ctx, inputs, Submission{Content: "test diff"})

	// Collective fails due to timeout, falls back to per-expert which also times out
	if len(result.Errors) == 0 {
		t.Error("expected errors from timeout, got none")
	}
}

func TestRunnerCollectiveErrorFallsBackToPerExpert(t *testing.T) {
	backend := &MockBackend{
		CollectiveErr: fmt.Errorf("API rate limited"),
		Results: map[string]ExpertVerdict{
			"kent-beck":      {Expert: "kent-beck", Verdict: VerdictPass, Confidence: 0.9},
			"bruce-schneier": {Expert: "bruce-schneier", Verdict: VerdictComment, Confidence: 0.8, Notes: []string{"Check auth"}},
		},
	}

	runner := &Runner{
		Backend: backend,
		Options: ReviewOptions{Concurrency: 2, Timeout: 10},
	}

	inputs := []ExpertInput{
		{Expert: &expert.Expert{ID: "kent-beck", Name: "Kent Beck", Focus: "TDD"}},
		{Expert: &expert.Expert{ID: "bruce-schneier", Name: "Bruce Schneier", Focus: "Security"}},
	}

	result := runner.Run(context.Background(), inputs, Submission{Content: "test diff"})

	if backend.collectiveCalls.Load() != 1 {
		t.Errorf("expected 1 collective call, got %d", backend.collectiveCalls.Load())
	}
	if backend.calls.Load() != 2 {
		t.Errorf("expected 2 per-expert fallback calls, got %d", backend.calls.Load())
	}
	if len(result.Perspectives) != 2 {
		t.Fatalf("expected 2 perspectives from fallback, got %d", len(result.Perspectives))
	}
	if result.Verdict != VerdictComment {
		t.Errorf("expected comment verdict from fallback, got %s", result.Verdict)
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors after successful fallback, got %v", result.Errors)
	}
}

func TestRunnerPerExpertPartialFailure(t *testing.T) {
	backend := &MockBackend{
		Results: map[string]ExpertVerdict{
			"kent-beck": {Expert: "kent-beck", Verdict: VerdictPass, Confidence: 0.9},
		},
		Errors: map[string]error{
			"bruce-schneier": fmt.Errorf("timeout"),
		},
	}

	runner := &Runner{
		Backend: backend,
		Options: ReviewOptions{Concurrency: 2, Timeout: 10},
	}

	// Single expert uses per-expert path
	inputs := []ExpertInput{
		{Expert: &expert.Expert{ID: "kent-beck", Name: "Kent Beck", Focus: "TDD"}},
	}

	result := runner.Run(context.Background(), inputs, Submission{Content: "test diff"})

	if len(result.Perspectives) != 1 {
		t.Errorf("expected 1 perspective, got %d", len(result.Perspectives))
	}
}

func TestRunnerPerExpertAllFail(t *testing.T) {
	backend := &MockBackend{
		Errors: map[string]error{
			"kent-beck": fmt.Errorf("timeout"),
		},
	}

	runner := &Runner{
		Backend: backend,
		Options: ReviewOptions{Concurrency: 2, Timeout: 10},
	}

	inputs := []ExpertInput{
		{Expert: &expert.Expert{ID: "kent-beck", Name: "Kent Beck", Focus: "TDD"}},
	}

	result := runner.Run(context.Background(), inputs, Submission{Content: "test diff"})

	if len(result.Perspectives) != 0 {
		t.Errorf("expected 0 perspectives, got %d", len(result.Perspectives))
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
}

func TestRunnerPerExpertContextCancellation(t *testing.T) {
	backend := &MockBackend{
		Delay: 5 * time.Second,
	}

	runner := &Runner{
		Backend: backend,
		Options: ReviewOptions{Concurrency: 2, Timeout: 1},
	}

	inputs := []ExpertInput{
		{Expert: &expert.Expert{ID: "kent-beck", Name: "Kent Beck", Focus: "TDD"}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result := runner.Run(ctx, inputs, Submission{Content: "test diff"})

	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error from cancellation, got %d errors", len(result.Errors))
	}
}

func TestRunnerPerExpertConcurrencyLimit(t *testing.T) {
	var maxConcurrent atomic.Int32
	var current atomic.Int32

	backend := &MockBackend{
		Delay: 50 * time.Millisecond,
	}

	wrapper := &concurrencyTracker{
		inner:         backend,
		current:       &current,
		maxConcurrent: &maxConcurrent,
	}

	runner := &Runner{
		Backend: wrapper,
		Options: ReviewOptions{Concurrency: 2, Timeout: 10},
	}

	// 5 experts triggers large-prompt fallback (collective prompt > threshold not guaranteed
	// with tiny bodies, so use large content to force fallback)
	largeContent := strings.Repeat("x", CollectiveThreshold)
	inputs := make([]ExpertInput, 5)
	for i := range inputs {
		inputs[i] = ExpertInput{
			Expert: &expert.Expert{ID: fmt.Sprintf("expert-%d", i), Name: fmt.Sprintf("Expert %d", i), Focus: "Testing"},
		}
	}

	runner.Run(context.Background(), inputs, Submission{Content: largeContent})

	if maxConcurrent.Load() > 2 {
		t.Errorf("max concurrent = %d, want <= 2", maxConcurrent.Load())
	}
}

type concurrencyTracker struct {
	inner         Backend
	current       *atomic.Int32
	maxConcurrent *atomic.Int32
}

func (c *concurrencyTracker) Review(ctx context.Context, e *expert.Expert, sub Submission) (ExpertVerdict, error) {
	n := c.current.Add(1)
	for {
		old := c.maxConcurrent.Load()
		if n <= old || c.maxConcurrent.CompareAndSwap(old, n) {
			break
		}
	}
	defer c.current.Add(-1)
	return c.inner.Review(ctx, e, sub)
}

func (c *concurrencyTracker) ReviewCollective(ctx context.Context, experts []*expert.Expert, sub Submission) (*SynthesizedResult, error) {
	return c.inner.ReviewCollective(ctx, experts, sub)
}
