package review

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/luuuc/council-cli/internal/expert"
)

// Backend defines the interface for executing a single expert review.
type Backend interface {
	Review(ctx context.Context, e *expert.Expert, sub Submission) (ExpertVerdict, error)
}

// CLIBackend spawns subprocess calls to an AI CLI for reviews.
type CLIBackend struct {
	Command string
	Args    []string
}

// knownCLIDefaults returns default args for known AI CLIs.
func knownCLIDefaults(command string) []string {
	base := command
	// Handle full paths: /usr/local/bin/claude -> claude
	if idx := strings.LastIndex(command, "/"); idx >= 0 {
		base = command[idx+1:]
	}

	switch base {
	case "claude":
		return []string{"-p", "--output-format", "text"}
	case "opencode":
		return []string{"-p"}
	default:
		return nil
	}
}

// NewCLIBackend creates a CLIBackend with sensible defaults for the given command.
// If args is nil or empty, defaults are applied for known CLIs.
func NewCLIBackend(command string, args []string) *CLIBackend {
	if len(args) == 0 {
		args = knownCLIDefaults(command)
	}
	return &CLIBackend{
		Command: command,
		Args:    args,
	}
}

// Review executes a single expert review via subprocess.
func (b *CLIBackend) Review(ctx context.Context, e *expert.Expert, sub Submission) (ExpertVerdict, error) {
	prompt := sub.RawPrompt
	if prompt == "" {
		prompt = BuildPrompt(e, sub)
	}

	// Build command args
	baseArgs := make([]string, len(b.Args))
	copy(baseArgs, b.Args)

	cmd := exec.CommandContext(ctx, b.Command, baseArgs...)

	// Use stdin for large prompts to avoid ARG_MAX limits (~256KB on most systems).
	// Threshold set conservatively below typical limits.
	const argMaxSafe = 128 * 1024
	if len(prompt) > argMaxSafe {
		cmd.Stdin = strings.NewReader(prompt)
	} else {
		cmd.Args = append(cmd.Args, prompt)
	}

	// CombinedOutput captures both stdout and stderr. Some CLIs write
	// review output to stderr in non-interactive mode.
	output, err := cmd.CombinedOutput()
	if err != nil {
		detail := ""
		if len(output) > 0 {
			detail = ": " + truncateBytes(output, 200)
		}
		return ExpertVerdict{}, fmt.Errorf("subprocess failed for %s%s: %w", e.ID, detail, err)
	}

	// RawPrompt mode: return the raw text directly instead of parsing verdict JSON.
	if sub.RawPrompt != "" {
		return ExpertVerdict{
			Expert:     e.ID,
			Verdict:    VerdictComment,
			Confidence: 1.0,
			Notes:      []string{strings.TrimSpace(string(output))},
		}, nil
	}

	verdict := ParseVerdict(e.ID, output)
	return verdict, nil
}

// truncateBytes returns a string of at most maxLen bytes from b.
func truncateBytes(b []byte, maxLen int) string {
	if len(b) <= maxLen {
		return string(b)
	}
	return string(b[:maxLen]) + "..."
}
