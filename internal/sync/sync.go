package sync

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/luuuc/council-cli/internal/config"
	"github.com/luuuc/council-cli/internal/expert"
)

// Target represents a sync target
type Target struct {
	Name     string
	Sync     func(experts []*expert.Expert, cfg *config.Config, dryRun bool) error
	Check    func() bool
	Location string
}

// Targets is the registry of available sync targets
var Targets = map[string]*Target{
	"claude": {
		Name:     "Claude Code",
		Location: ".claude/",
		Sync:     syncClaude,
		Check:    func() bool { return dirExists(".claude") },
	},
	"cursor": {
		Name:     "Cursor",
		Location: ".cursor/rules/ or .cursorrules",
		Sync:     syncCursor,
		Check:    func() bool { return dirExists(".cursor") || fileExists(".cursorrules") },
	},
	"windsurf": {
		Name:     "Windsurf",
		Location: ".windsurfrules",
		Sync:     syncWindsurf,
		Check:    func() bool { return fileExists(".windsurfrules") },
	},
	"generic": {
		Name:     "Generic",
		Location: "AGENTS.md",
		Sync:     syncGeneric,
		Check:    func() bool { return fileExists("AGENTS.md") },
	},
}

// SyncAll syncs to all configured targets
func SyncAll(cfg *config.Config, dryRun bool) error {
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
		if err := target.Sync(experts, cfg, dryRun); err != nil {
			return fmt.Errorf("failed to sync to %s: %w", targetName, err)
		}
	}

	return nil
}

// SyncTarget syncs to a specific target
func SyncTarget(targetName string, cfg *config.Config, dryRun bool) error {
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
	return target.Sync(experts, cfg, dryRun)
}

// Claude Code sync
func syncClaude(experts []*expert.Expert, cfg *config.Config, dryRun bool) error {
	// Create .claude/agents directory
	agentsDir := ".claude/agents"
	if !dryRun {
		if err := os.MkdirAll(agentsDir, 0755); err != nil {
			return err
		}
	}

	// Sync each expert as an agent file
	for _, e := range experts {
		path := filepath.Join(agentsDir, e.ID+".md")
		content := generateAgentFile(e)

		if dryRun {
			fmt.Printf("  Would create: %s\n", path)
		} else {
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return err
			}
			fmt.Printf("  Created: %s\n", path)
		}
	}

	// Create /council command if configured
	if cfg.Council.IncludeCouncilCommand {
		commandsDir := ".claude/commands"
		if !dryRun {
			if err := os.MkdirAll(commandsDir, 0755); err != nil {
				return err
			}
		}

		councilCmd := generateCouncilCommand(experts)
		path := filepath.Join(commandsDir, "council.md")

		if dryRun {
			fmt.Printf("  Would create: %s\n", path)
		} else {
			if err := os.WriteFile(path, []byte(councilCmd), 0644); err != nil {
				return err
			}
			fmt.Printf("  Created: %s\n", path)
		}
	}

	return nil
}

// Cursor sync
func syncCursor(experts []*expert.Expert, cfg *config.Config, dryRun bool) error {
	// Prefer .cursor/rules/ if .cursor exists, otherwise .cursorrules
	var path string
	if dirExists(".cursor") {
		rulesDir := ".cursor/rules"
		if !dryRun {
			if err := os.MkdirAll(rulesDir, 0755); err != nil {
				return err
			}
		}
		path = filepath.Join(rulesDir, "council.md")
	} else {
		path = ".cursorrules"
	}

	content := generateCombinedRules(experts)

	if dryRun {
		fmt.Printf("  Would create: %s\n", path)
	} else {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
		fmt.Printf("  Created: %s\n", path)
	}

	return nil
}

// Windsurf sync
func syncWindsurf(experts []*expert.Expert, cfg *config.Config, dryRun bool) error {
	content := generateCombinedRules(experts)
	path := ".windsurfrules"

	if dryRun {
		fmt.Printf("  Would create: %s\n", path)
	} else {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
		fmt.Printf("  Created: %s\n", path)
	}

	return nil
}

// Generic AGENTS.md sync
func syncGeneric(experts []*expert.Expert, cfg *config.Config, dryRun bool) error {
	content := generateAgentsMd(experts)
	path := "AGENTS.md"

	if dryRun {
		fmt.Printf("  Would create: %s\n", path)
	} else {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
		fmt.Printf("  Created: %s\n", path)
	}

	return nil
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
	tmpl := `# Code Review Council

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
`
	t, _ := template.New("council").Parse(tmpl)
	var buf bytes.Buffer
	t.Execute(&buf, experts)
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

		if len(e.Triggers.Paths) > 0 {
			parts = append(parts, fmt.Sprintf("- **Paths**: %s", strings.Join(e.Triggers.Paths, ", ")))
		}
		if len(e.Triggers.Keywords) > 0 {
			parts = append(parts, fmt.Sprintf("- **Keywords**: %s", strings.Join(e.Triggers.Keywords, ", ")))
		}

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

// Helper functions
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
