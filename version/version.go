package version

import (
	"fmt"
	"strings"
)

// validCharacters  is a list of characters valid in the appBuild string
const validCharacters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"

const (
	appMajor uint = 0
	appMinor uint = 1
	appPatch uint = 0
)

// appBuild is defined as a variable so it can be overridden during the build
// process with '-ldflags "-X github.com/kaspanet/kaspad/version.appBuild=foo"' if needed.
// It MUST only contain characters from validCharacters.
var appBuild string

var version = "" // string used for memoization of version

// Version returns the application version as a properly formed string
func Version() string {
	if version == "" {
		// Start with the major, minor, and patch versions.
		version = fmt.Sprintf("%d.%d.%d", appMajor, appMinor, appPatch)

		// Append build metadata if there is any. The build metadata
		// string is not appended if it contains invalid characters.
		checkAppBuild(appBuild)
		if appBuild != "" {
			version = fmt.Sprintf("%s-%s", version, appBuild)
		}
	}

	return version
}

// checkAppBuild returns the passed string unless it contains any characters not in validCharacters
// If any invalid characters are encountered - an empty string is returned
func checkAppBuild(appBuild string) {
	for _, r := range appBuild {
		if !strings.ContainsRune(validCharacters, r) {
			panic(fmt.Errorf("appBuild string (%s) contains forbidden characters. Only alphanumeric characters and dashes are allowed", appBuild))
		}
	}
}
