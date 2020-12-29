package flowcontext

import "github.com/kaspanet/kaspad/util/mstime"

const (
	syncRateWindowInMilliSeconds                         = 60_000
	syncRateMaxDeviation                                 = 0.05
	maxSelectedParentTimeDiffToAllowMiningInMilliSeconds = 300_000
)

func (f *FlowContext) updateRecentBlockAddedTimesWithLastBlock() {
	f.recentBlockAddedTimesMutex.Lock()
	defer f.recentBlockAddedTimesMutex.Unlock()

	f.removeOldBlockTimes()
	f.recentBlockAddedTimes = append(f.recentBlockAddedTimes, mstime.Now().UnixMilliseconds())
}

// removeOldBlockTimes removes from recentBlockAddedTimes block times
// older than syncRateWindowInMilliSeconds.
// This function is not safe for concurrent use.
func (f *FlowContext) removeOldBlockTimes() {
	now := mstime.Now().UnixMilliseconds()
	mostRecentBlockToKeep := 0
	for i, blockAddedTime := range f.recentBlockAddedTimes {
		if now-syncRateWindowInMilliSeconds < blockAddedTime {
			mostRecentBlockToKeep = i
			break
		}
	}
	f.recentBlockAddedTimes = f.recentBlockAddedTimes[mostRecentBlockToKeep:]
}

func (f *FlowContext) isSyncRateBelowMinimum() bool {
	f.recentBlockAddedTimesMutex.Lock()
	defer f.recentBlockAddedTimesMutex.Unlock()

	f.removeOldBlockTimes()

	now := mstime.Now().UnixMilliseconds()
	timeSinceStart := now - f.timeStarted
	if timeSinceStart <= syncRateWindowInMilliSeconds {
		return false
	}

	expectedBlocks := float64(syncRateWindowInMilliSeconds) / float64(f.cfg.NetParams().TargetTimePerBlock.Milliseconds())
	return 1-float64(len(f.recentBlockAddedTimes))/expectedBlocks > syncRateMaxDeviation
}

// ShouldMine returns whether it's ok to use block template from this node
// for mining purposes.
func (f *FlowContext) ShouldMine() (bool, error) {
	if f.isSyncRateBelowMinimum() {
		log.Debugf("The sync rate is below the minimum, so ShouldMine returns true")
		return true, nil
	}

	if f.IsIBDRunning() {
		log.Debugf("IBD is running, so ShouldMine returns false")
		return false, nil
	}

	virtualSelectedParent, err := f.domain.Consensus().GetVirtualSelectedParent()
	if err != nil {
		return false, err
	}

	virtualSelectedParentHeader, err := f.domain.Consensus().GetBlockHeader(virtualSelectedParent)
	if err != nil {
		return false, err
	}

	now := mstime.Now().UnixMilliseconds()
	if now-virtualSelectedParentHeader.TimeInMilliseconds() < maxSelectedParentTimeDiffToAllowMiningInMilliSeconds {
		log.Debugf("The selected tip timestamp is recent (%d), so ShouldMine returns true",
			virtualSelectedParentHeader.TimeInMilliseconds())
		return true, nil
	}

	log.Debugf("The selected tip timestamp is old (%d), so ShouldMine returns false",
		virtualSelectedParentHeader.TimeInMilliseconds())
	return false, nil
}
