package blockdag

import "time"

const syncRateWindowDuration = 15 * time.Minute

// addBlockProcessingTimestamp adds the last block processing timestamp in order to measure the recent sync rate.
//
// This function MUST be called with the DAG state lock held (for writes).
func (dag *BlockDAG) addBlockProcessingTimestamp() {
	now := time.Now()
	dag.recentBlockProcessingTimestamps = append(dag.recentBlockProcessingTimestamps, now)
	dag.recentBlockProcessingTimestamps = dag.recentBlockProcessingTimestampsRelevantWindow()
}

func (dag *BlockDAG) recentBlockProcessingTimestampsRelevantWindow() []time.Time {
	minTime := time.Now().Add(-syncRateWindowDuration)
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

	return dag.syncRate() < 1/dag.dagParams.TargetTimePerBlock.Seconds()*maxDeviation
}

func (dag *BlockDAG) uptime() time.Duration {
	return time.Now().Sub(dag.startTime)
}
