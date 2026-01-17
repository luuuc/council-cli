// Package fs provides common filesystem helper functions.
package fs

import (
	"os"
	"path/filepath"
)

// FileExists checks if a file or directory exists at the given path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a directory exists at the given path.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// FileExistsIn checks if a file exists within a directory.
func FileExistsIn(dir, name string) bool {
	return FileExists(filepath.Join(dir, name))
}

// DirExistsIn checks if a subdirectory exists within a directory.
func DirExistsIn(dir, name string) bool {
	return DirExists(filepath.Join(dir, name))
}

// ReadFile reads a file and returns its contents as a string.
// Returns an empty string if the file cannot be read.
func ReadFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// ReadFileIn reads a file within a directory and returns its contents as a string.
// Returns an empty string if the file cannot be read.
func ReadFileIn(dir, name string) string {
	return ReadFile(filepath.Join(dir, name))
}
