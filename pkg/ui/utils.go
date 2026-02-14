package ui

import "os"

// IsInteractive checks if stdin is a terminal (not a pipe)
func IsInteractive() bool {
	fileInfo, _ := os.Stdin.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
