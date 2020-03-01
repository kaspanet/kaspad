package locks

import (
	"fmt"
	"regexp"
	"runtime/debug"
)

func goroutineIDAndCallerToMutex(mutexFileName string) string {
	pattern := fmt.Sprintf("goroutine (\\d+)(.|\n)*%s.*\n(.+)\n(.+)", mutexFileName)
	re := regexp.MustCompile(pattern)
	stack := string(debug.Stack())
	matches := re.FindStringSubmatch(stack)
	goroutineID := matches[1]
	caller := matches[3]
	line := matches[4]
	return fmt.Sprintf("goroutine %s: %s ( %s )", goroutineID, caller, line)
}
