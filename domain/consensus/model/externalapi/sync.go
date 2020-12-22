package externalapi

// SyncInfo holds info about the current sync state of the consensus
type SyncInfo struct {
	IsAwaitingUTXOSet    bool
	IBDRootUTXOBlockHash *DomainHash
	HeaderCount          uint64
	BlockCount           uint64
}

// Clone returns a clone of SyncInfo
func (si *SyncInfo) Clone() *SyncInfo {
	return &SyncInfo{
		IsAwaitingUTXOSet:    si.IsAwaitingUTXOSet,
		IBDRootUTXOBlockHash: si.IBDRootUTXOBlockHash.Clone(),
		HeaderCount:          si.HeaderCount,
		BlockCount:           si.BlockCount,
	}
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal and Clone accordingly.
var _ = SyncInfo{false, &DomainHash{}, 0, 0}

// Equal returns whether si equals to other
func (si *SyncInfo) Equal(other *SyncInfo) bool {
	if si == nil || other == nil {
		return si == other
	}

	if si.IsAwaitingUTXOSet != other.IsAwaitingUTXOSet {
		return false
	}

	if !si.IBDRootUTXOBlockHash.Equal(other.IBDRootUTXOBlockHash) {
		return false
	}

	if si.HeaderCount != other.HeaderCount {
		return false
	}

	if si.BlockCount != other.BlockCount {
		return false
	}

	return true
}
