package creator

import (
	"os"
	"path/filepath"
)

const (
	// InstalledDir is the directory name for installed councils
	InstalledDir = "installed"
)

// BaseDir returns the base directory for council data.
// Uses os.UserConfigDir() for cross-platform support:
//   - macOS: ~/Library/Application Support/council/
//   - Linux: ~/.config/council/
//   - Windows: %AppData%\council\
func BaseDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "council"), nil
}

// InstalledPath returns the path to the installed councils directory.
func InstalledPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, InstalledDir), nil
}
