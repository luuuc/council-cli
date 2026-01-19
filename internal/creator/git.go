package creator

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// Cached git availability check
var (
	gitAvailableOnce   sync.Once
	gitAvailableResult bool
)

// GitOperations defines the interface for git operations.
// This allows for testing with mock implementations.
type GitOperations interface {
	Init() error
	IsRepo() bool
	Add(file string) error
	Commit(message string) error
	AddAndCommit(file, message string) error
	Status() (string, error)
	Log(limit int) ([]string, error)
	LogFile(file string, limit int) ([]string, error)
	RemoteAdd(name, url string) error
	RemoteGet(name string) (string, error)
	RemoteExists(name string) bool
	Push() error
	Pull() error
}

// GitRepo wraps git operations for the personal council.
type GitRepo struct {
	path string
}

// Ensure GitRepo implements GitOperations
var _ GitOperations = (*GitRepo)(nil)

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

// Init initializes a new git repository.
func (g *GitRepo) Init() error {
	if !gitAvailable() {
		return nil // Graceful degradation
	}
	return g.run("init")
}

// IsRepo checks if the path is a git repository.
func (g *GitRepo) IsRepo() bool {
	gitDir := filepath.Join(g.path, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// Add stages a file.
func (g *GitRepo) Add(file string) error {
	if !gitAvailable() || !g.IsRepo() {
		return nil
	}
	return g.run("add", file)
}

// Commit creates a commit with the given message.
func (g *GitRepo) Commit(message string) error {
	if !gitAvailable() || !g.IsRepo() {
		return nil
	}

	// Check if there are staged changes
	output, err := g.output("diff", "--cached", "--quiet")
	if err == nil && output == "" {
		// No staged changes, nothing to commit
		return nil
	}

	return g.run("commit", "-m", message)
}

// AddAndCommit stages and commits a file in one operation.
func (g *GitRepo) AddAndCommit(file, message string) error {
	if err := g.Add(file); err != nil {
		return err
	}
	return g.Commit(message)
}

// Status returns the git status output.
func (g *GitRepo) Status() (string, error) {
	if !gitAvailable() {
		return "", fmt.Errorf("git not installed")
	}
	if !g.IsRepo() {
		return "", fmt.Errorf("not a git repository")
	}
	return g.output("status", "--short")
}

// Log returns recent commit history.
func (g *GitRepo) Log(limit int) ([]string, error) {
	if !gitAvailable() {
		return nil, fmt.Errorf("git not installed")
	}
	if !g.IsRepo() {
		return nil, fmt.Errorf("not a git repository")
	}

	output, err := g.output("log", "--oneline", fmt.Sprintf("-n%d", limit))
	if err != nil {
		// Empty repo has no commits
		if strings.Contains(err.Error(), "does not have any commits") {
			return []string{}, nil
		}
		return nil, err
	}

	if output == "" {
		return []string{}, nil
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	return lines, nil
}

// LogFile returns commit history for a specific file.
func (g *GitRepo) LogFile(file string, limit int) ([]string, error) {
	if !gitAvailable() {
		return nil, fmt.Errorf("git not installed")
	}
	if !g.IsRepo() {
		return nil, fmt.Errorf("not a git repository")
	}

	output, err := g.output("log", "--oneline", fmt.Sprintf("-n%d", limit), "--", file)
	if err != nil {
		return nil, err
	}

	if output == "" {
		return []string{}, nil
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	return lines, nil
}

// RemoteAdd adds a remote.
func (g *GitRepo) RemoteAdd(name, url string) error {
	if !gitAvailable() {
		return fmt.Errorf("git not installed")
	}
	if !g.IsRepo() {
		return fmt.Errorf("not a git repository")
	}
	return g.run("remote", "add", name, url)
}

// RemoteGet returns the URL of a remote.
func (g *GitRepo) RemoteGet(name string) (string, error) {
	if !gitAvailable() {
		return "", fmt.Errorf("git not installed")
	}
	if !g.IsRepo() {
		return "", fmt.Errorf("not a git repository")
	}

	output, err := g.output("remote", "get-url", name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// RemoteExists checks if a remote exists.
func (g *GitRepo) RemoteExists(name string) bool {
	_, err := g.RemoteGet(name)
	return err == nil
}

// Push pushes to the remote.
func (g *GitRepo) Push() error {
	if !gitAvailable() {
		return fmt.Errorf("git not installed")
	}
	if !g.IsRepo() {
		return fmt.Errorf("not a git repository")
	}

	// First push needs to set upstream
	if err := g.run("push", "-u", "origin", "main"); err != nil {
		// Try master if main fails
		return g.run("push", "-u", "origin", "master")
	}
	return nil
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

// output executes a git command and returns its output.
func (g *GitRepo) output(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.path

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return "", fmt.Errorf("%s: %s", err, errMsg)
		}
		return "", err
	}

	return stdout.String(), nil
}
