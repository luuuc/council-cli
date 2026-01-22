package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// isInteractive returns true if stdin is a terminal (not piped).
func isInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// Confirm asks user for confirmation with a y/n prompt
func Confirm(prompt string) bool {
	fmt.Print(prompt + " [Y/n] ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "" || response == "y" || response == "yes"
}
