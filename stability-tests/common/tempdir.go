package common

import "io/ioutil"

// TempDir returns a temporary directory with the given pattern, prefixed with STABILITY_TEMP_DIR_
func TempDir(pattern string) (string, error) {
	const prefix = "STABILITY_TEMP_DIR_"
	return ioutil.TempDir("", prefix+pattern)
}
