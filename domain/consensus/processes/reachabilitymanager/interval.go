package reachabilitymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/pkg/errors"
	"math"
)

func newReachabilityInterval(start uint64, end uint64) *model.ReachabilityInterval {
	return &model.ReachabilityInterval{Start: start, End: end}
}

// intervalSize returns the size of this interval. Note that intervals are
// inclusive from both sides.
func intervalSize(ri *model.ReachabilityInterval) uint64 {
	return ri.End - ri.Start + 1
}

// intervalSplitInHalf splits this interval by a fraction of 0.5.
// See splitFraction for further details.
func intervalSplitInHalf(ri *model.ReachabilityInterval) (
	left *model.ReachabilityInterval, right *model.ReachabilityInterval, err error) {

	return intervalSplitFraction(ri, 0.5)
}

// intervalSplitFraction splits this interval to two parts such that their
// union is equal to the original interval and the first (left) part
// contains the given fraction of the original interval's size.
// Note: if the split results in fractional parts, this method rounds
// the first part up and the last part down.
func intervalSplitFraction(ri *model.ReachabilityInterval, fraction float64) (
	left *model.ReachabilityInterval, right *model.ReachabilityInterval, err error) {

	if fraction < 0 || fraction > 1 {
		return nil, nil, errors.Errorf("fraction must be between 0 and 1")
	}
	if intervalSize(ri) == 0 {
		return nil, nil, errors.Errorf("cannot split an empty interval")
	}

	allocationSize := uint64(math.Ceil(float64(intervalSize(ri)) * fraction))
	left = newReachabilityInterval(ri.Start, ri.Start+allocationSize-1)
	right = newReachabilityInterval(ri.Start+allocationSize, ri.End)
	return left, right, nil
}

// intervalSplitExact splits this interval to exactly |sizes| parts where
// |part_i| = sizes[i]. This method expects sum(sizes) to be exactly
// equal to the interval's size.
func intervalSplitExact(ri *model.ReachabilityInterval, sizes []uint64) ([]*model.ReachabilityInterval, error) {
	sizesSum := uint64(0)
	for _, size := range sizes {
		sizesSum += size
	}
	if sizesSum != intervalSize(ri) {
		return nil, errors.Errorf("sum of sizes must be equal to the interval's size")
	}

	intervals := make([]*model.ReachabilityInterval, len(sizes))
	start := ri.Start
	for i, size := range sizes {
		intervals[i] = newReachabilityInterval(start, start+size-1)
		start += size
	}
	return intervals, nil
}

// intervalSplitWithExponentialBias splits this interval to |sizes| parts
// by the allocation rule described below. This method expects sum(sizes)
// to be smaller or equal to the interval's size. Every part_i is
// allocated at least sizes[i] capacity. The remaining budget is
// split by an exponentially biased rule described below.
//
// This rule follows the GHOSTDAG protocol behavior where the child
// with the largest subtree is expected to dominate the competition
// for new blocks and thus grow the most. However, we may need to
// add slack for non-largest subtrees in order to make CPU reindexing
// attacks unworthy.
func intervalSplitWithExponentialBias(ri *model.ReachabilityInterval, sizes []uint64) ([]*model.ReachabilityInterval, error) {
	intervalSize := intervalSize(ri)
	sizesSum := uint64(0)
	for _, size := range sizes {
		sizesSum += size
	}
	if sizesSum > intervalSize {
		return nil, errors.Errorf("sum of sizes must be less than or equal to the interval's size")
	}
	if sizesSum == intervalSize {
		return intervalSplitExact(ri, sizes)
	}

	// Add a fractional bias to every size in the given sizes
	totalBias := intervalSize - sizesSum
	remainingBias := totalBias
	biasedSizes := make([]uint64, len(sizes))
	fractions := exponentialFractions(sizes)
	for i, fraction := range fractions {
		var bias uint64
		if i == len(fractions)-1 {
			bias = remainingBias
		} else {
			bias = uint64(math.Round(float64(totalBias) * fraction))
			if bias > remainingBias {
				bias = remainingBias
			}
		}
		biasedSizes[i] = sizes[i] + bias
		remainingBias -= bias
	}
	return intervalSplitExact(ri, biasedSizes)
}

// intervalContains returns true if ri contains other.
func intervalContains(ri *model.ReachabilityInterval, other *model.ReachabilityInterval) bool {
	return ri.Start <= other.Start && other.End <= ri.End
}
