package blockdag

import (
	"github.com/kaspanet/kaspad/util/mstime"
	"time"
)

const syncRateWindowDuration = 15 * time.Minute

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
