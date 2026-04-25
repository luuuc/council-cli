package review

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/luuuc/council/internal/expert"
)

// Runner orchestrates expert reviews.
type Runner struct {
	Backend Backend
	Options ReviewOptions
}

// ExpertInput pairs an expert with their blocking status from the pack.
type ExpertInput struct {
	Expert   *expert.Expert
	Blocking bool
}

// CollectiveThreshold is the byte-count threshold for the collective prompt.
// If the prompt exceeds this, the runner falls back to per-expert review.
// Default: 32KB (~8K tokens) — conservative for small-context models.
const CollectiveThreshold = 32 * 1024

// Run executes a collective review by default (one LLM call with all experts).
// Falls back to per-expert concurrent review when a single expert is specified
// or the estimated collective prompt exceeds CollectiveThreshold.
func (r *Runner) Run(ctx context.Context, inputs []ExpertInput, sub Submission) *SynthesizedResult {
	if len(inputs) == 1 {
		return r.runPerExpert(ctx, inputs, sub)
	}

	if estimateCollectiveSize(inputs, sub) > CollectiveThreshold {
		log.Println("collective prompt exceeds context threshold, falling back to per-expert review")
		return r.runPerExpert(ctx, inputs, sub)
	}

	return r.runCollective(ctx, inputs, sub)
}

// estimateCollectiveSize approximates the collective prompt size in bytes
// without building the full string. Sums expert content + submission + template overhead.
func estimateCollectiveSize(inputs []ExpertInput, sub Submission) int {
	const templateOverhead = 800
	size := templateOverhead + len(sub.Content) + len(sub.Context)
	for _, inp := range inputs {
		size += len(inp.Expert.Name) + len(inp.Expert.Focus) + len(inp.Expert.Body) + 20
	}
	return size
}

// runCollective executes a single collective LLM call.
// Falls back to per-expert review if the collective call fails.
func (r *Runner) runCollective(ctx context.Context, inputs []ExpertInput, sub Submission) *SynthesizedResult {
	timeout := time.Duration(r.Options.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}

	experts := make([]*expert.Expert, len(inputs))
	for i, inp := range inputs {
		experts[i] = inp.Expert
	}

	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := r.Backend.ReviewCollective(callCtx, experts, sub)
	if err != nil {
		log.Printf("collective review failed, falling back to per-expert: %s", err)
		return r.runPerExpert(ctx, inputs, sub)
	}

	// Override Blocking per perspective from pack config
	blockingByID := make(map[string]bool, len(inputs))
	for _, inp := range inputs {
		blockingByID[inp.Expert.ID] = inp.Blocking
	}
	for i := range result.Perspectives {
		result.Perspectives[i].Blocking = blockingByID[result.Perspectives[i].Expert]
	}

	// Validate overall verdict against hierarchy
	byID := make(map[string]*expert.Expert, len(experts))
	for _, e := range experts {
		byID[e.ID] = e
	}
	hierarchyVerdict := ResolveOverallVerdict(result.Perspectives, byID)
	if hierarchyVerdict.Severity() > result.Verdict.Severity() {
		result.Verdict = hierarchyVerdict
	}
	result.Blocking = ResolveBlocking(result.Perspectives)

	return result
}

// runPerExpert executes reviews in parallel with bounded concurrency (fallback path).
func (r *Runner) runPerExpert(ctx context.Context, inputs []ExpertInput, sub Submission) *SynthesizedResult {
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

			sem <- struct{}{}
			defer func() { <-sem }()

			expertCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			verdict, err := r.Backend.Review(expertCtx, inp.Expert, sub)
			if err != nil {
				results[idx] = result{
					err: fmt.Errorf("%s: %w", inp.Expert.ID, err),
				}
				return
			}

			verdict.Blocking = inp.Blocking
			results[idx] = result{verdict: verdict}
		}(i, input)
	}

	wg.Wait()

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
