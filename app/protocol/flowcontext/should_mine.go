package flowcontext

import "github.com/kaspanet/kaspad/util/mstime"

const (
	maxSelectedParentTimeDiffToAllowMiningInMilliSeconds = 60 * 60 * 1000 // 1 Hour
)

// ShouldMine returns whether it's ok to use block template from this node
// for mining purposes.
func (f *FlowContext) ShouldMine() (bool, error) {
	peers := f.Peers()
	if len(peers) == 0 {
		log.Debugf("The node is not connected, so ShouldMine returns false")
		return false, nil
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
