package review

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/luuuc/council-cli/internal/expert"
)

// Runner orchestrates parallel expert reviews.
type Runner struct {
	Backend Backend
	Options ReviewOptions
}

// ExpertInput pairs an expert with their blocking status from the pack.
type ExpertInput struct {
	Expert   *expert.Expert
	Blocking bool
}

// Run executes reviews for all experts in parallel with bounded concurrency.
// Failed reviews are recorded in errors; the review continues with remaining experts.
func (r *Runner) Run(ctx context.Context, inputs []ExpertInput, sub Submission) *SynthesizedResult {
	concurrency := r.Options.Concurrency
	if concurrency <= 0 {
		concurrency = DefaultConcurrency
	}

	timeout := time.Duration(r.Options.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}

	type result struct {
		verdict ExpertVerdict
		err     error
	}

	results := make([]result, len(inputs))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, input := range inputs {
		wg.Add(1)
		go func(idx int, inp ExpertInput) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Per-expert timeout
			expertCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			verdict, err := r.Backend.Review(expertCtx, inp.Expert, sub)
			if err != nil {
				results[idx] = result{
					err: fmt.Errorf("%s: %w", inp.Expert.ID, err),
				}
				return
			}

			// Set blocking from pack config
			verdict.Blocking = inp.Blocking
			results[idx] = result{verdict: verdict}
		}(i, input)
	}

	wg.Wait()

	// Collect verdicts and errors
	var verdicts []ExpertVerdict
	var errors []string
	experts := make([]*expert.Expert, 0, len(inputs))

	for i, r := range results {
		experts = append(experts, inputs[i].Expert)
		if r.err != nil {
			errors = append(errors, r.err.Error())
			continue
		}
		verdicts = append(verdicts, r.verdict)
	}

	return Synthesize(verdicts, experts, errors)
}
