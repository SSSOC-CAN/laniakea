package utils

import (
	"os"
)

// FileExists reports whether the named file or directory exists.
// This function is taken from https://github.com/lightningnetwork/lnd
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}