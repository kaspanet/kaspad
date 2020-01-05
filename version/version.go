package version

import (
	"bytes"
	"fmt"
	"strings"
)

// validAlphabet is a list of charecters valid in the appBuild string
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
		build := normalizeVerString(appBuild)
		if build != "" {
			version = fmt.Sprintf("%s-%s", version, build)
		}
	}

	return version
}

// normalizeVerString returns the passed string stripped of all characters which
// are not in validCharacters.
func normalizeVerString(str string) string {
	var result bytes.Buffer
	for _, r := range str {
		if strings.ContainsRune(validCharacters, r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}
