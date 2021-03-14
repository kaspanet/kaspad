// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package sorters

import "sort"

// Int64Slice implements sort.Interface to allow a slice of timestamps to
// be sorted.
type Int64Slice []int64

// Len returns the number of timestamps in the slice. It is part of the
// sort.Interface implementation.
func (s Int64Slice) Len() int {
	return len(s)
}

// Swap swaps the timestamps at the passed indices. It is part of the
// sort.Interface implementation.
func (s Int64Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less returns whether the timstamp with index i should sort before the
// timestamp with index j. It is part of the sort.Interface implementation.
func (s Int64Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

// Sort is a convenience method: s.Sort() calls sort.Sort(s).
func (s Int64Slice) Sort() { sort.Sort(s) }
