package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Version != 1 {
		t.Errorf("Default().Version = %d, want 1", cfg.Version)
	}
	// AI.Command should be empty (detected at runtime)
	if cfg.AI.Command != "" {
		t.Errorf("Default().AI.Command = %q, want empty (detected at runtime)", cfg.AI.Command)
	}
	if cfg.AI.Timeout != 120 {
		t.Errorf("Default().AI.Timeout = %d, want 120", cfg.AI.Timeout)
	}
	// Targets should be empty (detected at sync time)
	if len(cfg.Targets) != 0 {
		t.Errorf("Default().Targets length = %d, want 0 (detected at sync time)", len(cfg.Targets))
	}
}

func TestDetectAICommand(t *testing.T) {
	// Test with explicit command - should return it directly
	cfg := &Config{AI: AIConfig{Command: "myai"}}
	cmd, err := cfg.DetectAICommand()
	if err != nil {
		t.Errorf("DetectAICommand() with explicit command should not error: %v", err)
	}
	if cmd != "myai" {
		t.Errorf("DetectAICommand() = %q, want myai", cmd)
	}

	// Test with empty command - should detect or error
	cfg = &Config{}
	cmd, err = cfg.DetectAICommand()
	// Result depends on test environment - we just verify it doesn't panic
	// and returns a non-empty string if successful, or an error if not
	if err == nil && cmd == "" {
		t.Error("DetectAICommand() should return non-empty command or error")
	}
}

func TestPath(t *testing.T) {
	tests := []struct {
		parts []string
		want  string
	}{
		{[]string{}, ".council"},
		{[]string{"config.yaml"}, filepath.Join(".council", "config.yaml")},
		{[]string{"experts", "dhh.md"}, filepath.Join(".council", "experts", "dhh.md")},
	}

	for _, tt := range tests {
		got := Path(tt.parts...)
		if got != tt.want {
			t.Errorf("Path(%v) = %s, want %s", tt.parts, got, tt.want)
		}
	}
}

func TestExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	// Should not exist initially
	if Exists() {
		t.Error("Exists() should return false when .council doesn't exist")
	}

	// Create .council directory
	if err := os.MkdirAll(CouncilDir, 0755); err != nil {
		t.Fatalf("Failed to create council dir: %v", err)
	}

	if !Exists() {
		t.Error("Exists() should return true when .council exists")
	}

	// Create a file with same name (edge case)
	os.RemoveAll(CouncilDir)
	if err := os.WriteFile(CouncilDir, []byte("not a dir"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	if Exists() {
		t.Error("Exists() should return false when .council is a file, not directory")
	}
}

func TestLoadAndSave(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	// Load should fail when not initialized
	_, err = Load()
	if err == nil {
		t.Error("Load() should error when council not initialized")
	}

	// Create .council directory and save config
	if err := os.MkdirAll(CouncilDir, 0755); err != nil {
		t.Fatalf("Failed to create council dir: %v", err)
	}

	cfg := Default()
	cfg.AI.Command = "aichat"
	cfg.Targets = []string{"claude", "windsurf"}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load should succeed now
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.AI.Command != "aichat" {
		t.Errorf("Load().AI.Command = %s, want aichat", loaded.AI.Command)
	}
	if len(loaded.Targets) != 2 {
		t.Errorf("Load().Targets length = %d, want 2", len(loaded.Targets))
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.MkdirAll(CouncilDir, 0755); err != nil {
		t.Fatalf("Failed to create council dir: %v", err)
	}

	// Write invalid YAML
	invalidYAML := []byte("version: [invalid\n  yaml: content")
	if err := os.WriteFile(Path(ConfigFile), invalidYAML, 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	_, err = Load()
	if err == nil {
		t.Error("Load() should error on invalid YAML")
	}
}

func TestConstants(t *testing.T) {
	if CouncilDir != ".council" {
		t.Errorf("CouncilDir = %s, want .council", CouncilDir)
	}
	if ConfigFile != "config.yaml" {
		t.Errorf("ConfigFile = %s, want config.yaml", ConfigFile)
	}
	if ExpertsDir != "experts" {
		t.Errorf("ExpertsDir = %s, want experts", ExpertsDir)
	}
	if CommandsDir != "commands" {
		t.Errorf("CommandsDir = %s, want commands", CommandsDir)
	}
}
