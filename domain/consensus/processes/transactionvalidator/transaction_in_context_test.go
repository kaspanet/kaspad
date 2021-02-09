package transactionvalidator

import (
	"github.com/kaspanet/kaspad/util/mstime"
	"testing"
)

// TestSequenceLocksActive tests the SequenceLockActive function to ensure it
// works as expected in all possible combinations/scenarios.
func TestSequenceLocksActive(t *testing.T) {
	tests := []struct {
		seqLock        sequenceLock
		blockBlueScore uint64
		mtp            mstime.Time

		want bool
	}{
		// Block based sequence lock with equal block blue score.
		{seqLock: sequenceLock{-1, 1000}, blockBlueScore: 1001, mtp: mstime.UnixMilliseconds(9), want: true},

		// Time based sequence lock with mtp past the absolute time.
		{seqLock: sequenceLock{30, -1}, blockBlueScore: 2, mtp: mstime.UnixMilliseconds(31), want: true},

		// Block based sequence lock with current blue score below seq lock block blue score.
		{seqLock: sequenceLock{-1, 1000}, blockBlueScore: 90, mtp: mstime.UnixMilliseconds(9), want: false},

		// Time based sequence lock with current time before lock time.
		{seqLock: sequenceLock{30, -1}, blockBlueScore: 2, mtp: mstime.UnixMilliseconds(29), want: false},

		// Block based sequence lock at the same blue score, so shouldn't yet be active.
		{seqLock: sequenceLock{-1, 1000}, blockBlueScore: 1000, mtp: mstime.UnixMilliseconds(9), want: false},

		// Time based sequence lock with current time equal to lock time, so shouldn't yet be active.
		{seqLock: sequenceLock{30, -1}, blockBlueScore: 2, mtp: mstime.UnixMilliseconds(30), want: false},
	}

	validator := transactionValidator{}
	for i, test := range tests {
		got := validator.sequenceLockActive(&test.seqLock, test.blockBlueScore, test.mtp.UnixMilliseconds())
		if got != test.want {
			t.Fatalf("SequenceLockActive #%d got %v want %v", i, got, test.want)
		}
	}
}
