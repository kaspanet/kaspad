package flowcontext

import "github.com/kaspanet/kaspad/util/mstime"

// IsNearlySynced returns whether this node is considered synced or close to being synced. This info
// is used to determine if it's ok to use a block template from this node for mining purposes.
func (f *FlowContext) IsNearlySynced() (bool, error) {
	peers := f.Peers()
	if len(peers) == 0 {
		log.Debugf("The node is not connected to peers, so IsNearlySynced returns false")
		return false, nil
	}

	virtualSelectedParent, err := f.domain.Consensus().GetVirtualSelectedParent()
	if err != nil {
		return false, err
	}

	if virtualSelectedParent.Equal(f.Config().NetParams().GenesisHash) {
		return false, nil
	}

	virtualSelectedParentHeader, err := f.domain.Consensus().GetBlockHeader(virtualSelectedParent)
	if err != nil {
		return false, err
	}

	now := mstime.Now().UnixMilliseconds()
	// As a heuristic, we allow the node to mine if he is likely to be within the current DAA window of fully synced nodes.
	// Such blocks contribute to security by maintaining the current difficulty despite possibly being slightly out of sync.
	if now-virtualSelectedParentHeader.TimeInMilliseconds() < f.expectedDAAWindowDurationInMilliseconds {
		log.Debugf("The selected tip timestamp is recent (%d), so IsNearlySynced returns true",
			virtualSelectedParentHeader.TimeInMilliseconds())
		return true, nil
	}

	log.Debugf("The selected tip timestamp is old (%d), so IsNearlySynced returns false",
		virtualSelectedParentHeader.TimeInMilliseconds())
	return false, nil
}
