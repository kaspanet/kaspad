package blockdag

import (
	"github.com/kaspanet/kaspad/util/mstime"
	"time"
)

const (
	syncRateWindowDuration = 15 * time.Minute

	// isDAGCurrentMaxDiff is the number of blocks from the network tips (estimated by timestamps) for the current
	// to be considered not synced
	isDAGCurrentMaxDiff = 40_000
)

// addBlockProcessingTimestamp adds the last block processing timestamp in order to measure the recent sync rate.
//
// This function MUST be called with the DAG state lock held (for writes).
func (dag *BlockDAG) addBlockProcessingTimestamp() {
	now := mstime.Now()
	dag.recentBlockProcessingTimestamps = append(dag.recentBlockProcessingTimestamps, now)
	dag.removeNonRecentTimestampsFromRecentBlockProcessingTimestamps()
}

// removeNonRecentTimestampsFromRecentBlockProcessingTimestamps removes timestamps older than syncRateWindowDuration
// from dag.recentBlockProcessingTimestamps
//
// This function MUST be called with the DAG state lock held (for writes).
func (dag *BlockDAG) removeNonRecentTimestampsFromRecentBlockProcessingTimestamps() {
	dag.recentBlockProcessingTimestamps = dag.recentBlockProcessingTimestampsRelevantWindow()
}

func (dag *BlockDAG) recentBlockProcessingTimestampsRelevantWindow() []mstime.Time {
	minTime := mstime.Now().Add(-syncRateWindowDuration)
	windowStartIndex := len(dag.recentBlockProcessingTimestamps)
	for i, processTime := range dag.recentBlockProcessingTimestamps {
		if processTime.After(minTime) {
			windowStartIndex = i
			break
		}
	}
	return dag.recentBlockProcessingTimestamps[windowStartIndex:]
}

// syncRate returns the rate of processed
// blocks in the last syncRateWindowDuration
// duration.
func (dag *BlockDAG) syncRate() float64 {
	dag.RLock()
	defer dag.RUnlock()
	return float64(len(dag.recentBlockProcessingTimestampsRelevantWindow())) / syncRateWindowDuration.Seconds()
}

// IsSyncRateBelowThreshold checks whether the sync rate
// is below the expected threshold.
func (dag *BlockDAG) IsSyncRateBelowThreshold(maxDeviation float64) bool {
	if dag.uptime() < syncRateWindowDuration {
		return false
	}

	return dag.syncRate() < 1/dag.Params.TargetTimePerBlock.Seconds()*maxDeviation
}

func (dag *BlockDAG) uptime() time.Duration {
	return mstime.Now().Sub(dag.startTime)
}

// isSynced returns whether or not the DAG believes it is synced. Several
// factors are used to guess, but the key factors that allow the DAG to
// believe it is synced are:
//  - Latest block has a timestamp newer than 24 hours ago
//
// This function MUST be called with the DAG state lock held (for reads).
func (dag *BlockDAG) isSynced() bool {
	// Not synced if the virtual's selected parent has a timestamp
	// before 24 hours ago. If the DAG is empty, we take the genesis
	// block timestamp.
	//
	// The DAG appears to be syncned if none of the checks reported
	// otherwise.
	var dagTimestamp int64
	selectedTip := dag.selectedTip()
	if selectedTip == nil {
		dagTimestamp = dag.Params.GenesisBlock.Header.Timestamp.UnixMilliseconds()
	} else {
		dagTimestamp = selectedTip.Timestamp
	}
	dagTime := mstime.UnixMilliseconds(dagTimestamp)
	return dag.Now().Sub(dagTime) <= isDAGCurrentMaxDiff*dag.Params.TargetTimePerBlock
}

// IsSynced returns whether or not the DAG believes it is synced. Several
// factors are used to guess, but the key factors that allow the DAG to
// believe it is synced are:
//  - Latest block has a timestamp newer than 24 hours ago
//
// This function is safe for concurrent access.
func (dag *BlockDAG) IsSynced() bool {
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()

	return dag.isSynced()
}
