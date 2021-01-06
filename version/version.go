package version

import (
	"fmt"
	"strings"
)

// validCharacters  is a list of characters valid in the appBuild string
const validCharacters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"

const (
	appMajor uint = 0
	appMinor uint = 8
	appPatch uint = 4
)

// appBuild is defined as a variable so it can be overridden during the build
// process with '-ldflags "-X github.com/kaspanet/kaspad/version.appBuild=foo"' if needed.
// It MUST only contain characters from validCharacters.
var appBuild string

var version = "" // string used for memoization of version

func init() {
	if version == "" {
		// Start with the major, minor, and patch versions.
		version = fmt.Sprintf("%d.%d.%d", appMajor, appMinor, appPatch)

		// Append build metadata if there is any.
		// Panic if any invalid characters are encountered
		if appBuild != "" {
			checkAppBuild(appBuild)

			version = fmt.Sprintf("%s-%s", version, appBuild)
		}
	}
}

// Version returns the application version as a properly formed string
func Version() string {
	return version
}

// checkAppBuild verifies that appBuild does not contain any characters outside of validCharacters.
// In case of any invalid characters checkAppBuild panics
func checkAppBuild(appBuild string) {
	for _, r := range appBuild {
		if !strings.ContainsRune(validCharacters, r) {
			panic(fmt.Errorf("appBuild string (%s) contains forbidden characters. Only alphanumeric characters and dashes are allowed", appBuild))
		}
	}
}
