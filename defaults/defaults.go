// Package defaults get settings.
package defaults

import (
	"os"
	"runtime"
)

// UserHomeDir returns user home directory by OS.
func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}

	return os.Getenv("HOME")
}
