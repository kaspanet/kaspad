package locks

import (
	"fmt"
	"regexp"
	"runtime/debug"
)

func goroutineIDAndCallerToMutex() string {
	re := regexp.MustCompile("goroutine (\\d+)(.+\n){7}(.+)\n(.+)")
	stack := string(debug.Stack())
	matches := re.FindStringSubmatch(stack)
	goroutineID := matches[1]
	caller := matches[3]
	line := matches[4]
	return fmt.Sprintf("goroutine %s: %s ( %s )", goroutineID, caller, line)
}
