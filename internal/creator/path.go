package creator

import (
	"os"
	"path/filepath"
)

const (
	// MyCouncilDir is the directory name for the personal council
	MyCouncilDir = "my-council"

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

// MyCouncilPath returns the path to the personal council directory.
func MyCouncilPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, MyCouncilDir), nil
}

// InstalledPath returns the path to the installed councils directory.
func InstalledPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, InstalledDir), nil
}

// PersonaPath returns the full path to a persona file in my-council.
func PersonaPath(id string) (string, error) {
	myCouncil, err := MyCouncilPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(myCouncil, id+".md"), nil
}

// Initialized checks if the personal council has been initialized.
func Initialized() bool {
	path, err := MyCouncilPath()
	if err != nil {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
