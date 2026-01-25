package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// Cached git availability check
var (
	gitAvailableOnce   sync.Once
	gitAvailableResult bool
)

// GitRepo wraps git operations for installed repositories.
type GitRepo struct {
	path string
}

// NewGitRepo creates a GitRepo for the given path.
func NewGitRepo(path string) *GitRepo {
	return &GitRepo{path: path}
}

// gitAvailable checks if git is installed (cached).
func gitAvailable() bool {
	gitAvailableOnce.Do(func() {
		_, err := exec.LookPath("git")
		gitAvailableResult = err == nil
	})
	return gitAvailableResult
}

// IsRepo checks if the path is a git repository.
func (g *GitRepo) IsRepo() bool {
	gitDir := filepath.Join(g.path, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// Pull pulls from the remote.
func (g *GitRepo) Pull() error {
	if !gitAvailable() {
		return fmt.Errorf("git not installed")
	}
	if !g.IsRepo() {
		return fmt.Errorf("not a git repository")
	}
	return g.run("pull")
}

// run executes a git command.
func (g *GitRepo) run(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
