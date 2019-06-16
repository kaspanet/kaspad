package util

import "sort"

// SearchSlice uses binary search to find and return the smallest index i
// in [0, n) at which f(i) is true, assuming that on the range [0, n),
// f(i) == true implies f(i+1) == true. That is, Search requires that
// f is false for some (possibly empty) prefix of the input range [0, n)
// and then true for the (possibly empty) remainder; Search returns
// the first true index.
// Search calls f(i) only for i in the range [0, n).
func SearchSlice(sliceLength int, searchFunc func(int) bool) (foundIndex int, ok bool) {
	result := sort.Search(sliceLength, searchFunc)
	if result == sliceLength {
		return -1, false
	}
	return result, true
}
