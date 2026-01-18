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

//go:embed templates/council.md.tmpl
var councilCommandTemplateStr string

//go:embed templates/council-add.md
var councilAddCommand string

//go:embed templates/council-detect.md
var councilDetectCommand string

// Pre-compiled template for council command generation
var councilCommandTemplate = template.Must(template.New("council").Parse(councilCommandTemplateStr))

// allCommands defines all available slash commands and their templates
// Co-located here so adding a new command requires updating one place
var allCommands = map[string]string{
	"council":        "", // Uses councilCommandTemplate (dynamic, needs experts)
	"council-add":    councilAddCommand,
	"council-detect": councilDetectCommand,
}

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

	// Clean up stale files if requested
	if opts.Clean {
		// Remove stale command files
		for cmd := range allCommands {
			if !cfg.Council.HasCommand(cmd) {
				path := filepath.Join(commandsDir, cmd+".md")
				if err := removeFile(path, opts.DryRun); err != nil {
					return err
				}
			}
		}

		// Remove stale agent files (experts no longer in .council/experts/)
		if err := cleanStaleAgents(agentsDir, experts, opts.DryRun); err != nil {
			return err
		}
	}

	return nil
}

// cleanStaleAgents removes agent files that no longer have corresponding experts
func cleanStaleAgents(agentsDir string, experts []*expert.Expert, dryRun bool) error {
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Build set of current expert IDs
	currentIDs := make(map[string]bool)
	for _, e := range experts {
		currentIDs[e.ID] = true
	}

	// Remove files for experts that no longer exist
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".md")
		if !currentIDs[id] {
			path := filepath.Join(agentsDir, entry.Name())
			if err := removeFile(path, dryRun); err != nil {
				return err
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

	// Clean up stale files if requested
	if opts.Clean {
		// Remove stale command files (OpenCode only supports council-add and council-detect)
		openCodeCommands := []string{"council-add", "council-detect"}
		for _, cmd := range openCodeCommands {
			if !cfg.Council.HasCommand(cmd) {
				path := filepath.Join(agentDir, cmd+".md")
				if err := removeFile(path, opts.DryRun); err != nil {
					return err
				}
			}
		}

		// Remove stale agent files
		if err := cleanStaleAgentsOpenCode(agentDir, experts, openCodeCommands, opts.DryRun); err != nil {
			return err
		}
	}

	return nil
}

// cleanStaleAgentsOpenCode removes agent files that no longer have corresponding experts
// It excludes command files (council-add, council-detect) from cleanup
func cleanStaleAgentsOpenCode(agentDir string, experts []*expert.Expert, commandFiles []string, dryRun bool) error {
	entries, err := os.ReadDir(agentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Build set of current expert IDs
	currentIDs := make(map[string]bool)
	for _, e := range experts {
		currentIDs[e.ID] = true
	}

	// Build set of command file names to exclude
	commandSet := make(map[string]bool)
	for _, cmd := range commandFiles {
		commandSet[cmd] = true
	}

	// Remove files for experts that no longer exist
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".md")
		// Skip command files
		if commandSet[id] {
			continue
		}
		if !currentIDs[id] {
			path := filepath.Join(agentDir, entry.Name())
			if err := removeFile(path, dryRun); err != nil {
				return err
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

