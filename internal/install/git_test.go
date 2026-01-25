package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewGitRepo(t *testing.T) {
	repo := NewGitRepo("/some/path")
	if repo == nil {
		t.Fatal("NewGitRepo() returned nil")
	}
	if repo.path != "/some/path" {
		t.Errorf("NewGitRepo() path = %q, want /some/path", repo.path)
	}
}

func TestGitAvailable(t *testing.T) {
	// This test just verifies the function doesn't panic
	// The result depends on whether git is installed on the system
	result := gitAvailable()
	// Just verify it's a boolean (true or false)
	if result != true && result != false {
		t.Error("gitAvailable() should return a boolean")
	}
}

func TestGitRepo_IsRepo_WithFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git as a file (not a directory)
	gitFile := filepath.Join(tmpDir, ".git")
	if err := os.WriteFile(gitFile, []byte("gitdir: ../worktree/.git"), 0644); err != nil {
		t.Fatal(err)
	}

	repo := NewGitRepo(tmpDir)

	// .git as a file should return false (worktree case not supported by simple check)
	if repo.IsRepo() {
		t.Error("IsRepo() = true when .git is a file, want false")
	}
}

func TestGitRepo_IsRepo_WithSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")
	gitDir := filepath.Join(targetDir, ".git")

	// Create actual .git directory in target
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create symlink
	linkDir := filepath.Join(tmpDir, "link")
	if err := os.Symlink(targetDir, linkDir); err != nil {
		t.Fatal(err)
	}

	repo := NewGitRepo(linkDir)

	if !repo.IsRepo() {
		t.Error("IsRepo() = false for symlinked directory with .git, want true")
	}
}

func TestGitRepo_Pull_ErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		expectedErr string
	}{
		{
			name: "not a git repository",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			expectedErr: "not a git repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			repo := NewGitRepo(path)

			err := repo.Pull()
			if err == nil {
				t.Error("Pull() should return error")
				return
			}
			if err.Error() != tt.expectedErr {
				t.Errorf("Pull() error = %q, want %q", err.Error(), tt.expectedErr)
			}
		})
	}
}
