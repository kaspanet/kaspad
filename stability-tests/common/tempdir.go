package common

import "os"

// TempDir returns a temporary directory with the given pattern, prefixed with STABILITY_TEMP_DIR_
func TempDir(pattern string) (string, error) {
	const prefix = "STABILITY_TEMP_DIR_"
	return os.MkdirTemp("", prefix+pattern)
}
