package sync

import (
	"os"
	"strings"
	"testing"

	"github.com/luuuc/council-cli/internal/config"
	"github.com/luuuc/council-cli/internal/expert"
	"github.com/luuuc/council-cli/internal/fs"
)

func TestGenerateCouncilCommand(t *testing.T) {
	experts := []*expert.Expert{
		{
			ID:    "kent-beck",
			Name:  "Kent Beck",
			Focus: "Test-driven development",
		},
		{
			ID:    "dhh",
			Name:  "DHH",
			Focus: "Rails and productivity",
		},
	}

	result := generateCouncilCommand(experts)

	// Check for key elements
	if !strings.Contains(result, "Code Review Council") {
		t.Error("generateCouncilCommand() missing title")
	}
	if !strings.Contains(result, "$ARGUMENTS") {
		t.Error("generateCouncilCommand() missing $ARGUMENTS placeholder")
	}
	if !strings.Contains(result, "Kent Beck") {
		t.Error("generateCouncilCommand() missing first expert name")
	}
	if !strings.Contains(result, "DHH") {
		t.Error("generateCouncilCommand() missing second expert name")
	}
	if !strings.Contains(result, "Test-driven development") {
		t.Error("generateCouncilCommand() missing first expert focus")
	}
}

func TestGenerateCouncilCommand_EmptyExperts(t *testing.T) {
	// Test with empty expert list - should not panic
	experts := []*expert.Expert{}

	result := generateCouncilCommand(experts)

	// Should still have the header and instructions
	if !strings.Contains(result, "Code Review Council") {
		t.Error("generateCouncilCommand() should have title even with empty experts")
	}
	if !strings.Contains(result, "Instructions") {
		t.Error("generateCouncilCommand() should have instructions even with empty experts")
	}
}

func TestGenerateCouncilCommand_SpecialCharacters(t *testing.T) {
	// Test with special characters that might cause template issues
	experts := []*expert.Expert{
		{
			ID:    "special",
			Name:  "Expert with <html> & \"quotes\"",
			Focus: "Testing {{templates}} and $variables",
		},
	}

	result := generateCouncilCommand(experts)

	// Should not panic and should contain the special characters
	if !strings.Contains(result, "<html>") {
		t.Error("generateCouncilCommand() should preserve special characters")
	}
}

func TestGenerateAgentsMd(t *testing.T) {
	experts := []*expert.Expert{
		{
			ID:         "expert-1",
			Name:       "Expert One",
			Focus:      "Focus one",
			Philosophy: "Philosophy here.",
			Principles: []string{"Principle 1"},
		},
	}

	result := generateAgentsMd(experts)

	if !strings.Contains(result, "AGENTS.md") {
		t.Error("generateAgentsMd() missing AGENTS.md header")
	}
	if !strings.Contains(result, "Expert One") {
		t.Error("generateAgentsMd() missing expert name")
	}
	if !strings.Contains(result, "expert-1") {
		t.Error("generateAgentsMd() missing expert ID")
	}
	if !strings.Contains(result, "Philosophy here.") {
		t.Error("generateAgentsMd() missing philosophy")
	}
}

func TestSyncClaude(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "council-sync-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	// Create council structure with an expert
	_ = os.MkdirAll(config.Path(config.ExpertsDir), 0755)
	testExpert := &expert.Expert{
		ID:    "test",
		Name:  "Test Expert",
		Focus: "Testing",
	}
	_ = testExpert.Save()

	cfg := config.Default()
	experts := []*expert.Expert{testExpert}

	// Test dry run
	err = syncClaude(experts, cfg, Options{DryRun: true})
	if err != nil {
		t.Errorf("syncClaude() dry run error = %v", err)
	}

	// Verify nothing was created in dry run
	if _, err := os.Stat(".claude/agents"); !os.IsNotExist(err) {
		t.Error("syncClaude() dry run should not create directories")
	}

	// Test actual sync
	err = syncClaude(experts, cfg, Options{DryRun: false})
	if err != nil {
		t.Errorf("syncClaude() error = %v", err)
	}

	// Verify agent file was created
	agentPath := ".claude/agents/test.md"
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		t.Errorf("syncClaude() did not create agent file at %s", agentPath)
	}

	// Verify council command was created
	commandPath := ".claude/commands/council.md"
	if _, err := os.Stat(commandPath); os.IsNotExist(err) {
		t.Errorf("syncClaude() did not create council command at %s", commandPath)
	}

	// Read and verify council command content
	content, _ := os.ReadFile(commandPath)
	if !strings.Contains(string(content), "Test Expert") {
		t.Error("Council command should contain expert name")
	}
}

func TestSyncGeneric(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "council-sync-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cfg := config.Default()
	experts := []*expert.Expert{
		{
			ID:    "test",
			Name:  "Test Expert",
			Focus: "Testing",
		},
	}

	err = syncGeneric(experts, cfg, Options{DryRun: false})
	if err != nil {
		t.Errorf("syncGeneric() error = %v", err)
	}

	if _, err := os.Stat("AGENTS.md"); os.IsNotExist(err) {
		t.Error("syncGeneric() should create AGENTS.md")
	}

	content, _ := os.ReadFile("AGENTS.md")
	if !strings.Contains(string(content), "Test Expert") {
		t.Error("AGENTS.md should contain expert name")
	}
}

func TestSyncAllNoExperts(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "council-sync-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	// Create empty council structure
	_ = os.MkdirAll(config.Path(config.ExpertsDir), 0755)

	cfg := config.Default()

	err = SyncAll(cfg, Options{DryRun: false})
	if err == nil {
		t.Error("SyncAll() should error when no experts exist")
	}
}

func TestTargetsRegistry(t *testing.T) {
	// Verify all expected targets are registered
	expectedTargets := []string{"claude", "generic", "opencode"}

	for _, name := range expectedTargets {
		target, ok := Targets[name]
		if !ok {
			t.Errorf("Target %s not found in registry", name)
			continue
		}
		if target.Name == "" {
			t.Errorf("Target %s has empty Name", name)
		}
		if target.Location == "" {
			t.Errorf("Target %s has empty Location", name)
		}
		if target.Sync == nil {
			t.Errorf("Target %s has nil Sync function", name)
		}
		if target.Check == nil {
			t.Errorf("Target %s has nil Check function", name)
		}
	}
}

func TestFileExistsAndDirExists(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "council-sync-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	// Test non-existent
	if fs.FileExists("nonexistent.txt") {
		t.Error("fs.FileExists() should return false for non-existent file")
	}
	if fs.DirExists("nonexistent") {
		t.Error("fs.DirExists() should return false for non-existent directory")
	}

	// Create a file and directory
	_ = os.WriteFile("test.txt", []byte("content"), 0644)
	_ = os.MkdirAll("testdir", 0755)

	if !fs.FileExists("test.txt") {
		t.Error("fs.FileExists() should return true for existing file")
	}
	if !fs.DirExists("testdir") {
		t.Error("fs.DirExists() should return true for existing directory")
	}

	// File should not be detected as directory
	if fs.DirExists("test.txt") {
		t.Error("fs.DirExists() should return false for file")
	}
}

func TestGenerateAgentFile(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "council-sync-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	// Create council structure with an expert
	_ = os.MkdirAll(config.Path(config.ExpertsDir), 0755)
	testExpert := &expert.Expert{
		ID:    "test",
		Name:  "Test Expert",
		Focus: "Testing",
		Body:  "# Test Expert\n\nCustom body content.",
	}
	_ = testExpert.Save()

	// Test generateAgentFile reads from disk
	result := generateAgentFile(testExpert)

	if !strings.Contains(result, "Test Expert") {
		t.Error("generateAgentFile() should contain expert name")
	}
	if !strings.Contains(result, "Custom body content") {
		t.Error("generateAgentFile() should contain custom body")
	}
}

func TestSyncOpenCode(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "council-sync-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cfg := config.Default()
	experts := []*expert.Expert{
		{
			ID:         "test",
			Name:       "Test Expert",
			Focus:      "Testing",
			Philosophy: "Test philosophy.",
			Principles: []string{"Principle 1"},
			RedFlags:   []string{"Red flag 1"},
		},
	}

	// Test dry run
	err = syncOpenCode(experts, cfg, Options{DryRun: true})
	if err != nil {
		t.Errorf("syncOpenCode() dry run error = %v", err)
	}

	// Verify nothing was created in dry run
	if _, err := os.Stat(".opencode/agent"); !os.IsNotExist(err) {
		t.Error("syncOpenCode() dry run should not create directories")
	}

	// Test actual sync
	err = syncOpenCode(experts, cfg, Options{DryRun: false})
	if err != nil {
		t.Errorf("syncOpenCode() error = %v", err)
	}

	// Verify agent file was created
	agentPath := ".opencode/agent/test.md"
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		t.Errorf("syncOpenCode() did not create agent file at %s", agentPath)
	}

	// Read and verify content
	content, _ := os.ReadFile(agentPath)
	contentStr := string(content)

	// Verify OpenCode-specific frontmatter format
	if !strings.Contains(contentStr, "description: Testing") {
		t.Error("OpenCode agent should have description in frontmatter")
	}
	if !strings.Contains(contentStr, "mode: subagent") {
		t.Error("OpenCode agent should have mode: subagent in frontmatter")
	}
	if !strings.Contains(contentStr, "Test Expert") {
		t.Error("OpenCode agent should contain expert name")
	}
	if !strings.Contains(contentStr, "Test philosophy.") {
		t.Error("OpenCode agent should contain philosophy")
	}
	if !strings.Contains(contentStr, "Principle 1") {
		t.Error("OpenCode agent should contain principles")
	}
	if !strings.Contains(contentStr, "Red flag 1") {
		t.Error("OpenCode agent should contain red flags")
	}
}

func TestGenerateOpenCodeAgent(t *testing.T) {
	e := &expert.Expert{
		ID:         "kent-beck",
		Name:       "Kent Beck",
		Focus:      "TDD and clean code",
		Philosophy: "Test-driven development leads to better design.",
		Principles: []string{"Red-green-refactor", "Simple design"},
		RedFlags:   []string{"No tests", "Complex mocking"},
	}

	result := generateOpenCodeAgent(e)

	// Verify frontmatter structure
	if !strings.HasPrefix(result, "---\n") {
		t.Error("generateOpenCodeAgent() should start with YAML frontmatter delimiter")
	}
	if !strings.Contains(result, "description: TDD and clean code") {
		t.Error("generateOpenCodeAgent() should have description matching focus")
	}
	if !strings.Contains(result, "mode: subagent") {
		t.Error("generateOpenCodeAgent() should have mode: subagent")
	}

	// Verify body content
	if !strings.Contains(result, "# Kent Beck") {
		t.Error("generateOpenCodeAgent() should have expert name as heading")
	}
	if !strings.Contains(result, "channeling Kent Beck") {
		t.Error("generateOpenCodeAgent() should have channeling intro")
	}
	if !strings.Contains(result, "## Philosophy") {
		t.Error("generateOpenCodeAgent() should have Philosophy section")
	}
	if !strings.Contains(result, "## Principles") {
		t.Error("generateOpenCodeAgent() should have Principles section")
	}
	if !strings.Contains(result, "## Red Flags") {
		t.Error("generateOpenCodeAgent() should have Red Flags section")
	}
	if !strings.Contains(result, "## Review Style") {
		t.Error("generateOpenCodeAgent() should have Review Style section")
	}
}

func TestSyncTargetUnknown(t *testing.T) {
	cfg := config.Default()
	err := SyncTarget("unknown-target", cfg, Options{DryRun: false})
	if err == nil {
		t.Error("SyncTarget() should error for unknown target")
	}
	if !strings.Contains(err.Error(), "unknown target") {
		t.Errorf("Error should mention 'unknown target', got: %v", err)
	}
}
