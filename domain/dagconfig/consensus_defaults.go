package dagconfig

import (
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
)

// The documentation refers to the following constants which aren't explicated in the code:
//	d - an upper bound on the round trip time of a block
//	delta - the expected fraction of time the width of the network exceeds defaultGHOSTDAGK
//
// For more information about defaultGHOSTDAGK, and its dependency on delta and defaultTargetTimePerBlock
// please refer to the PHANTOM paper: https://eprint.iacr.org/2018/104.pdf
//
// For more information about the DAA constants defaultDifficultyAdjustmentWindowSize, defaultTimestampDeviationTolerance,
// and their relation to defaultGHOSTDAGK and defaultTargetTimePerBlock see:
// https://research.kas.pa/t/handling-timestamp-manipulations/97
//
// For more information about defaultMergeSetSizeLimit, defaultFinalityDuration and their relation to pruning, see:
// https://research.kas.pa/t/a-proposal-for-finality-in-ghostdag/66/17
// https://research.kas.pa/t/some-of-the-intuition-behind-the-design-of-the-invalidation-rules-for-pruning/95
//

const (
	defaultMaxCoinbasePayloadLength = 150
	// defaultMaxBlockSize is a bound on the size of a block in bytes, larger values increase the bound d
	// on the round trip time of a block, which affects the other parameters as described below
	defaultMaxBlockSize = 1_000_000
	// defaultMaxBlockParents is the number of blocks any block can point to.
	// Should be about d/defaultTargetTimePerBlock where d is a bound on the round trip time of a block.
	defaultMaxBlockParents         = 20
	defaultMassPerTxByte           = 1
	defaultMassPerScriptPubKeyByte = 10
	defaultMassPerSigOp            = 10000
	// defaultMergeSetSizeLimit is a bound on the size of the past of a block and the size of the past
	// of its selected parent. Any block which violates this bound is invalid.
	// Should be at least an order of magnitude smaller than defaultFinalityDuration/defaultTargetTimePerBlock.
	// (Higher values make pruning attacks easier by a constant, lower values make merging after a split or a spike
	// in block take longer)
	defaultMergeSetSizeLimit                       = 1000
	defaultMaxMassAcceptedByBlock                  = 10000000
	defaultBaseSubsidy                             = 50 * constants.SompiPerKaspa
	defaultCoinbasePayloadScriptPublicKeyMaxLength = 150
	// defaultGHOSTDAGK is a bound on the number of blue blocks in the anticone of a blue block. Approximates the maximal
	// width of the network.
	// Formula (1) in section 4.2 of the PHATOM paper shows how to calculate defaultGHOSTDAGK. The delta term represents a bound
	// on the expected fraction of the network life in which the width was higher than defaultGHOSTDAGK. The current value of K
	// was calculated for d = 5 seconds and delta = 0.05.
	defaultGHOSTDAGK = 18
	// defaultDifficultyAdjustmentWindowSize is the number of blocks in a block's past used to calculate its difficulty
	// target.
	defaultDifficultyAdjustmentWindowSize = 2640
	// defaultTimestampDeviationTolerance is the allowed deviance of an inconming block's timestamp, measured in block delays.
	// A new block can't hold a timestamp lower than the median timestamp of the (defaultTimestampDeviationTolerance*2-1) blocks
	// with highest accumulated blue work in its past, such blocks are considered invalid.
	// A new block can't hold a timestamp higher than the local system time + defaultTimestampDeviationTolerance/defaultTargetTimePerBlock,
	// such blocks are not marked as invalid but are rejected.
	defaultTimestampDeviationTolerance = 132
	// defaultFinalityDuration is an approximate lower bound of how old the finality block is. The finality block is chosen to
	// be the newest block in the selected chain whose blue score difference from the selected tip is at least
	// defaultFinalityDuration/defaultTargetTimePerBlock.
	// The pruning block is selected similarly, with the following duration:
	//	pruning block duration =
	//		2*defaultFinalityDuration/defaultTargetTimePerBlock + 4*defaultMergeSetSizeLimit*defaultGHOSTDAGK + 2*defaultGHOSTDAGK + 2
	defaultFinalityDuration = 24 * time.Hour
	// defaultTargetTimePerBlock represents how much time should pass on average between two consecutive block creations.
	// Should be parametrized such that the average width of the DAG is about defaultMaxBlockParents and such that most of the
	// time the width of the DAG is at most defaultGHOSTDAGK.
	defaultTargetTimePerBlock = 1 * time.Second
)
