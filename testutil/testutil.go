package testutil

import "fmt"

// AreErrorsEqual returns whether errors have the same type
// and same error string from .Error().
func AreErrorsEqual(err1, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	}
	if err1 == nil && err2 != nil {
		return false
	}
	if fmt.Sprintf("%T", err1) != fmt.Sprintf("%T", err2) {
		return false
	}
	return err1.Error() == err2.Error()
}
