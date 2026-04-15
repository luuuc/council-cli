package review

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/luuuc/council-cli/internal/expert"
)

// MockBackend returns canned verdicts for testing.
type MockBackend struct {
	Results map[string]ExpertVerdict
	Errors  map[string]error
	Delay   time.Duration
	calls   atomic.Int32
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

func TestRunnerHappyPath(t *testing.T) {
	backend := &MockBackend{
		Results: map[string]ExpertVerdict{
			"kent-beck": {
				Expert: "kent-beck", Verdict: VerdictComment,
				Confidence: 0.8, Notes: []string{"Add test"},
			},
			"bruce-schneier": {
				Expert: "bruce-schneier", Verdict: VerdictPass,
				Confidence: 0.95,
			},
		},
	}

	runner := &Runner{
		Backend: backend,
		Options: ReviewOptions{Concurrency: 2, Timeout: 10},
	}

	inputs := []ExpertInput{
		{Expert: &expert.Expert{ID: "kent-beck", Name: "Kent Beck", Focus: "TDD"}, Blocking: false},
		{Expert: &expert.Expert{ID: "bruce-schneier", Name: "Bruce Schneier", Focus: "Security"}, Blocking: true},
	}

	result := runner.Run(context.Background(), inputs, Submission{Content: "test diff"})

	if len(result.Perspectives) != 2 {
		t.Fatalf("expected 2 perspectives, got %d", len(result.Perspectives))
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got %v", result.Errors)
	}
	if result.Verdict != VerdictComment {
		t.Errorf("expected comment verdict, got %s", result.Verdict)
	}
}

func TestRunnerPartialFailure(t *testing.T) {
	backend := &MockBackend{
		Results: map[string]ExpertVerdict{
			"kent-beck": {
				Expert: "kent-beck", Verdict: VerdictPass, Confidence: 0.9,
			},
		},
		Errors: map[string]error{
			"bruce-schneier": fmt.Errorf("timeout"),
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

	if len(result.Perspectives) != 1 {
		t.Errorf("expected 1 perspective, got %d", len(result.Perspectives))
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
}

func TestRunnerAllFail(t *testing.T) {
	backend := &MockBackend{
		Errors: map[string]error{
			"kent-beck":      fmt.Errorf("timeout"),
			"bruce-schneier": fmt.Errorf("OOM"),
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

	if len(result.Perspectives) != 0 {
		t.Errorf("expected 0 perspectives, got %d", len(result.Perspectives))
	}
	if len(result.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(result.Errors))
	}
}

func TestRunnerContextCancellation(t *testing.T) {
	backend := &MockBackend{
		Delay: 5 * time.Second,
	}

	runner := &Runner{
		Backend: backend,
		Options: ReviewOptions{Concurrency: 2, Timeout: 1}, // 1s timeout
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

func TestRunnerConcurrencyLimit(t *testing.T) {
	var maxConcurrent atomic.Int32
	var current atomic.Int32

	backend := &MockBackend{
		Delay: 50 * time.Millisecond,
	}

	// Wrap to track concurrency
	wrapper := &concurrencyTracker{
		inner:         backend,
		current:       &current,
		maxConcurrent: &maxConcurrent,
	}

	runner := &Runner{
		Backend: wrapper,
		Options: ReviewOptions{Concurrency: 2, Timeout: 10},
	}

	inputs := make([]ExpertInput, 5)
	for i := range inputs {
		inputs[i] = ExpertInput{
			Expert: &expert.Expert{ID: fmt.Sprintf("expert-%d", i), Name: fmt.Sprintf("Expert %d", i), Focus: "Testing"},
		}
	}

	runner.Run(context.Background(), inputs, Submission{Content: "test"})

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
