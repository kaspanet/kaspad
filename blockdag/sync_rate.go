package blockdag

import "time"

const syncRateWindowDuration = 15 * time.Minute

// addBlockProcessTime adds the last the last
// block process time in order to measure the
// recent sync rate.
//
// This function MUST be called with the DAG state lock held (for writes).
func (dag *BlockDAG) addBlockProcessTime() {
	now := time.Now()
	dag.recentBlockProcessTimes = append(dag.recentBlockProcessTimes, now)
	dag.recentBlockProcessTimes = dag.recentBlockProcessTimesRelevantWindow()
}

func (dag *BlockDAG) recentBlockProcessTimesRelevantWindow() []time.Time {
	minTime := time.Now().Add(-syncRateWindowDuration)
	windowStartIndex := len(dag.recentBlockProcessTimes)
	for i, processTime := range dag.recentBlockProcessTimes {
		if processTime.After(minTime) {
			windowStartIndex = i
			break
		}
	}
	return dag.recentBlockProcessTimes[windowStartIndex:]
}

// syncRate returns the rate of processed
// blocks in the last syncRateWindowDuration
// duration.
func (dag *BlockDAG) syncRate() float64 {
	dag.RLock()
	defer dag.RUnlock()
	return float64(len(dag.recentBlockProcessTimesRelevantWindow())) / syncRateWindowDuration.Seconds()
}

// IsSyncRateBelowThreshold checks whether the sync rate
// is below an expected threshold.
func (dag *BlockDAG) IsSyncRateBelowThreshold() bool {
	if dag.uptime() < syncRateWindowDuration {
		return false
	}

	const maxDeviation = 0.05
	return dag.syncRate() < 1/dag.dagParams.TargetTimePerBlock.Seconds()*(1-maxDeviation)
}

func (dag *BlockDAG) uptime() time.Duration {
	return time.Now().Sub(dag.startTime)
}
