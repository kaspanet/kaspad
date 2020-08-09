package blockdag

// BehaviorFlags is a bitmask defining tweaks to the normal behavior when
// performing DAG processing and consensus rules checks.
type BehaviorFlags uint32

const (
	// BFFastAdd may be set to indicate that several checks can be avoided
	// for the block since it is already known to fit into the DAG due to
	// already proving it correct links into the DAG.
	BFFastAdd BehaviorFlags = 1 << iota

	// BFNoPoWCheck may be set to indicate the proof of work check which
	// ensures a block hashes to a value less than the required target will
	// not be performed.
	BFNoPoWCheck

	// BFWasUnorphaned may be set to indicate that a block was just now
	// unorphaned
	BFWasUnorphaned

	// BFAfterDelay may be set to indicate that a block had timestamp too far
	// in the future, just finished the delay
	BFAfterDelay

	// BFIsSync may be set to indicate that the block was sent as part of the
	// netsync process
	BFIsSync

	// BFWasStored is set to indicate that the block was previously stored
	// in the block index but was never fully processed
	BFWasStored

	// BFDisallowDelay is set to indicate that a delayed block should be rejected.
	// This is used for the case where a block is submitted through RPC.
	BFDisallowDelay

	// BFDisallowOrphans is set to indicate that an orphan block should be rejected.
	// This is used for the case where a block is submitted through RPC.
	BFDisallowOrphans

	// BFNone is a convenience value to specifically indicate no flags.
	BFNone BehaviorFlags = 0
)
