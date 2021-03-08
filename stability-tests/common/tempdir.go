package common

import "io/ioutil"

func TempDir(pattern string) (string, error) {
	const prefix = "STABILITY_TEMP_DIR_"
	return ioutil.TempDir("", prefix+pattern)
}
