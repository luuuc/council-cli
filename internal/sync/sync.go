package sync

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/luuuc/council-cli/internal/config"
	"github.com/luuuc/council-cli/internal/expert"
	"github.com/luuuc/council-cli/internal/fs"
)

//go:embed templates/council-add.md
var councilAddCommand string

//go:embed templates/council-detect.md
var councilDetectCommand string

// Pre-compiled template for council command generation
var councilCommandTemplate = template.Must(template.New("council").Parse(`# Code Review Council

Convene the council to review: $ARGUMENTS

## Council Members

{{range .}}
### {{.Name}}
**Focus**: {{.Focus}}
{{end}}

## Instructions

Review the code from each expert's perspective. For each expert:
1. State the expert's name
2. Provide their assessment focused on their domain
3. Note any concerns or suggestions

At the end, synthesize the key points and provide actionable recommendations.
`))

// Options configures sync behavior
type Options struct {
	DryRun bool // Show what would be done without making changes
	Clean  bool // Remove stale files not in current config
}

// Target represents a sync target
type Target struct {
	Name     string
	Sync     func(experts []*expert.Expert, cfg *config.Config, opts Options) error
	Check    func() bool
	Location string
}

// Targets is the registry of available sync targets
var Targets = map[string]*Target{
	"claude": {
		Name:     "Claude Code",
		Location: ".claude/",
		Sync:     syncClaude,
		Check:    func() bool { return fs.DirExists(".claude") },
	},
	"cursor": {
		Name:     "Cursor",
		Location: ".cursor/rules/ or .cursorrules",
		Sync:     syncCursor,
		Check:    func() bool { return fs.DirExists(".cursor") || fs.FileExists(".cursorrules") },
	},
	"windsurf": {
		Name:     "Windsurf",
		Location: ".windsurfrules",
		Sync:     syncWindsurf,
		Check:    func() bool { return fs.FileExists(".windsurfrules") },
	},
	"generic": {
		Name:     "Generic",
		Location: "AGENTS.md",
		Sync:     syncGeneric,
		Check:    func() bool { return fs.FileExists("AGENTS.md") },
	},
	"opencode": {
		Name:     "OpenCode",
		Location: ".opencode/agent/",
		Sync:     syncOpenCode,
		Check:    func() bool { return fs.DirExists(".opencode") || fs.FileExists("opencode.json") },
	},
}

// SyncAll syncs to all configured targets
func SyncAll(cfg *config.Config, opts Options) error {
	experts, err := expert.List()
	if err != nil {
		return err
	}

	if len(experts) == 0 {
		return fmt.Errorf("no experts to sync - add some with 'council add' or 'council setup --apply'")
	}

	for _, targetName := range cfg.Targets {
		target, ok := Targets[targetName]
		if !ok {
			fmt.Printf("Warning: unknown target '%s', skipping\n", targetName)
			continue
		}

		fmt.Printf("Syncing to %s (%s)...\n", target.Name, target.Location)
		if err := target.Sync(experts, cfg, opts); err != nil {
			return fmt.Errorf("failed to sync to %s: %w", targetName, err)
		}
	}

	return nil
}

// writeFile writes content to path, or prints what would be written in dry-run mode
func writeFile(path, content string, dryRun bool) error {
	if dryRun {
		fmt.Printf("  Would create: %s\n", path)
		return nil
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Printf("  Created: %s\n", path)
	return nil
}

// removeFile removes a file if it exists, or prints what would be removed in dry-run mode
func removeFile(path string, dryRun bool) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to do
	}
	if dryRun {
		fmt.Printf("  Would remove: %s\n", path)
		return nil
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	fmt.Printf("  Removed: %s\n", path)
	return nil
}

// SyncTarget syncs to a specific target
func SyncTarget(targetName string, cfg *config.Config, opts Options) error {
	target, ok := Targets[targetName]
	if !ok {
		return fmt.Errorf("unknown target: %s", targetName)
	}

	experts, err := expert.List()
	if err != nil {
		return err
	}

	if len(experts) == 0 {
		return fmt.Errorf("no experts to sync")
	}

	fmt.Printf("Syncing to %s (%s)...\n", target.Name, target.Location)
	return target.Sync(experts, cfg, opts)
}

// allCommands is the list of all possible council commands
var allCommands = []string{"council", "council-add", "council-detect"}

// Claude Code sync
func syncClaude(experts []*expert.Expert, cfg *config.Config, opts Options) error {
	// Create .claude/agents directory
	agentsDir := ".claude/agents"
	if !opts.DryRun {
		if err := os.MkdirAll(agentsDir, 0755); err != nil {
			return err
		}
	}

	// Sync each expert as an agent file
	for _, e := range experts {
		path := filepath.Join(agentsDir, e.ID+".md")
		if err := writeFile(path, generateAgentFile(e), opts.DryRun); err != nil {
			return err
		}
	}

	// Create commands directory if any commands are enabled
	hasCommands := len(cfg.Council.Commands) > 0
	commandsDir := ".claude/commands"
	if hasCommands && !opts.DryRun {
		if err := os.MkdirAll(commandsDir, 0755); err != nil {
			return err
		}
	}

	// Create /council command if configured
	if cfg.Council.HasCommand("council") {
		path := filepath.Join(commandsDir, "council.md")
		if err := writeFile(path, generateCouncilCommand(experts), opts.DryRun); err != nil {
			return err
		}
	}

	// Create /council-add command if configured
	if cfg.Council.HasCommand("council-add") {
		path := filepath.Join(commandsDir, "council-add.md")
		if err := writeFile(path, councilAddCommand, opts.DryRun); err != nil {
			return err
		}
	}

	// Create /council-detect command if configured
	if cfg.Council.HasCommand("council-detect") {
		path := filepath.Join(commandsDir, "council-detect.md")
		if err := writeFile(path, councilDetectCommand, opts.DryRun); err != nil {
			return err
		}
	}

	// Clean up stale command files if requested
	if opts.Clean {
		for _, cmd := range allCommands {
			if !cfg.Council.HasCommand(cmd) {
				path := filepath.Join(commandsDir, cmd+".md")
				if err := removeFile(path, opts.DryRun); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Cursor sync
func syncCursor(experts []*expert.Expert, cfg *config.Config, opts Options) error {
	// Prefer .cursor/rules/ if .cursor exists, otherwise .cursorrules
	var path string
	if fs.DirExists(".cursor") {
		rulesDir := ".cursor/rules"
		if !opts.DryRun {
			if err := os.MkdirAll(rulesDir, 0755); err != nil {
				return err
			}
		}
		path = filepath.Join(rulesDir, "council.md")
	} else {
		path = ".cursorrules"
	}

	return writeFile(path, generateCombinedRules(experts), opts.DryRun)
}

// Windsurf sync
func syncWindsurf(experts []*expert.Expert, cfg *config.Config, opts Options) error {
	return writeFile(".windsurfrules", generateCombinedRules(experts), opts.DryRun)
}

// Generic AGENTS.md sync
func syncGeneric(experts []*expert.Expert, cfg *config.Config, opts Options) error {
	return writeFile("AGENTS.md", generateAgentsMd(experts), opts.DryRun)
}

// OpenCode sync
func syncOpenCode(experts []*expert.Expert, cfg *config.Config, opts Options) error {
	// Create .opencode/agent directory
	agentDir := ".opencode/agent"
	if !opts.DryRun {
		if err := os.MkdirAll(agentDir, 0755); err != nil {
			return err
		}
	}

	// Sync each expert as an agent file
	for _, e := range experts {
		path := filepath.Join(agentDir, e.ID+".md")
		if err := writeFile(path, generateOpenCodeAgent(e), opts.DryRun); err != nil {
			return err
		}
	}

	// Create /council-add command if configured
	if cfg.Council.HasCommand("council-add") {
		path := filepath.Join(agentDir, "council-add.md")
		if err := writeFile(path, generateOpenCodeCommand("Add expert to council with AI-generated content", councilAddCommand), opts.DryRun); err != nil {
			return err
		}
	}

	// Create /council-detect command if configured
	if cfg.Council.HasCommand("council-detect") {
		path := filepath.Join(agentDir, "council-detect.md")
		if err := writeFile(path, generateOpenCodeCommand("Detect stack and suggest experts", councilDetectCommand), opts.DryRun); err != nil {
			return err
		}
	}

	// Clean up stale command files if requested
	if opts.Clean {
		// OpenCode only supports council-add and council-detect
		openCodeCommands := []string{"council-add", "council-detect"}
		for _, cmd := range openCodeCommands {
			if !cfg.Council.HasCommand(cmd) {
				path := filepath.Join(agentDir, cmd+".md")
				if err := removeFile(path, opts.DryRun); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// generateOpenCodeCommand creates OpenCode command file content
func generateOpenCodeCommand(description, body string) string {
	var parts []string
	parts = append(parts, "---")
	parts = append(parts, fmt.Sprintf("description: %s", description))
	parts = append(parts, "mode: subagent")
	parts = append(parts, "---")
	parts = append(parts, "")
	parts = append(parts, body)
	return strings.Join(parts, "\n")
}

// generateOpenCodeAgent creates OpenCode agent file content
func generateOpenCodeAgent(e *expert.Expert) string {
	var parts []string

	// OpenCode uses different frontmatter format
	parts = append(parts, "---")
	parts = append(parts, fmt.Sprintf("description: %s", e.Focus))
	parts = append(parts, "mode: subagent")
	parts = append(parts, "---")
	parts = append(parts, "")
	parts = append(parts, fmt.Sprintf("# %s", e.Name))
	parts = append(parts, "")
	parts = append(parts, fmt.Sprintf("You are channeling %s, known for expertise in %s.", e.Name, e.Focus))
	parts = append(parts, "")

	if e.Philosophy != "" {
		parts = append(parts, "## Philosophy")
		parts = append(parts, "")
		parts = append(parts, strings.TrimSpace(e.Philosophy))
		parts = append(parts, "")
	}

	if len(e.Principles) > 0 {
		parts = append(parts, "## Principles")
		parts = append(parts, "")
		for _, p := range e.Principles {
			parts = append(parts, fmt.Sprintf("- %s", p))
		}
		parts = append(parts, "")
	}

	if len(e.RedFlags) > 0 {
		parts = append(parts, "## Red Flags")
		parts = append(parts, "")
		parts = append(parts, "Watch for these patterns:")
		for _, r := range e.RedFlags {
			parts = append(parts, fmt.Sprintf("- %s", r))
		}
		parts = append(parts, "")
	}

	parts = append(parts, "## Review Style")
	parts = append(parts, "")
	parts = append(parts, "When reviewing code, focus on your area of expertise. Be direct and specific.")
	parts = append(parts, "Explain your reasoning. Suggest concrete improvements.")

	return strings.Join(parts, "\n")
}

// generateAgentFile creates Claude Code agent file content
func generateAgentFile(e *expert.Expert) string {
	// Read the original expert file and return its content
	data, err := os.ReadFile(e.Path())
	if err != nil {
		// Fallback to regenerating
		return fmt.Sprintf("---\nid: %s\nname: %s\nfocus: %s\n---\n\n%s", e.ID, e.Name, e.Focus, e.Body)
	}
	return string(data)
}

// generateCouncilCommand creates the /council slash command
func generateCouncilCommand(experts []*expert.Expert) string {
	var buf bytes.Buffer
	if err := councilCommandTemplate.Execute(&buf, experts); err != nil {
		// Fallback to simple format if template fails
		return "# Code Review Council\n\nConvene the council to review: $ARGUMENTS\n"
	}
	return buf.String()
}

// generateCombinedRules creates combined rules for Cursor/Windsurf
func generateCombinedRules(experts []*expert.Expert) string {
	var parts []string

	parts = append(parts, "# Expert Council")
	parts = append(parts, "")
	parts = append(parts, "This project uses an expert council pattern for code review guidance.")
	parts = append(parts, "")

	for _, e := range experts {
		parts = append(parts, fmt.Sprintf("## %s", e.Name))
		parts = append(parts, fmt.Sprintf("**Focus**: %s", e.Focus))
		parts = append(parts, "")

		if e.Philosophy != "" {
			parts = append(parts, strings.TrimSpace(e.Philosophy))
			parts = append(parts, "")
		}

		if len(e.Principles) > 0 {
			parts = append(parts, "**Principles:**")
			for _, p := range e.Principles {
				parts = append(parts, fmt.Sprintf("- %s", p))
			}
			parts = append(parts, "")
		}

		if len(e.RedFlags) > 0 {
			parts = append(parts, "**Watch for:**")
			for _, r := range e.RedFlags {
				parts = append(parts, fmt.Sprintf("- %s", r))
			}
			parts = append(parts, "")
		}
	}

	return strings.Join(parts, "\n")
}

// generateAgentsMd creates AGENTS.md content
func generateAgentsMd(experts []*expert.Expert) string {
	var parts []string

	parts = append(parts, "# AGENTS.md - Expert Council")
	parts = append(parts, "")
	parts = append(parts, "This file defines expert personas for AI coding assistants.")
	parts = append(parts, "")
	parts = append(parts, "## Council Members")
	parts = append(parts, "")

	for _, e := range experts {
		parts = append(parts, fmt.Sprintf("### %s", e.Name))
		parts = append(parts, fmt.Sprintf("- **ID**: %s", e.ID))
		parts = append(parts, fmt.Sprintf("- **Focus**: %s", e.Focus))
		parts = append(parts, "")

		if e.Philosophy != "" {
			parts = append(parts, strings.TrimSpace(e.Philosophy))
			parts = append(parts, "")
		}

		if len(e.Principles) > 0 {
			parts = append(parts, "**Principles:**")
			for _, p := range e.Principles {
				parts = append(parts, fmt.Sprintf("- %s", p))
			}
			parts = append(parts, "")
		}
	}

	return strings.Join(parts, "\n")
}

