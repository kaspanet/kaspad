package externalapi

import "fmt"

// Each of the following represent one of the possible sync
// states of the consensus
const (
	SyncStateRelay SyncState = iota
	SyncStateMissingGenesis
	SyncStateHeadersFirst
	SyncStateMissingUTXOSet
	SyncStateMissingBlockBodies
)

// SyncState represents the current sync state of the consensus
type SyncState uint8

func (s SyncState) String() string {
	switch s {
	case SyncStateRelay:
		return "SyncStateRelay"
	case SyncStateMissingGenesis:
		return "SyncStateMissingGenesis"
	case SyncStateHeadersFirst:
		return "SyncStateHeadersFirst"
	case SyncStateMissingUTXOSet:
		return "SyncStateMissingUTXOSet"
	case SyncStateMissingBlockBodies:
		return "SyncStateMissingBlockBodies"
	}

	return fmt.Sprintf("<unknown state (%d)>", s)
}

// Clone returns a clone of SyncState
func (s SyncState) Clone() SyncState {
	return s
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal accordingly.
var _ SyncState = 0

// Equal returns whether si equals to other
func (s SyncState) Equal(other SyncState) bool {
	return s == other
}

// SyncInfo holds info about the current sync state of the consensus
type SyncInfo struct {
	State                SyncState
	IBDRootUTXOBlockHash *DomainHash
	HeaderCount          uint64
	BlockCount           uint64
}

// Clone returns a clone of SyncInfo
func (si *SyncInfo) Clone() *SyncInfo {
	return &SyncInfo{
		State:                si.State.Clone(),
		IBDRootUTXOBlockHash: si.IBDRootUTXOBlockHash.Clone(),
		HeaderCount:          si.HeaderCount,
		BlockCount:           si.BlockCount,
	}
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal accordingly.
var _ = SyncInfo{SyncState(0), &DomainHash{}, 0, 0}

// Equal returns whether si equals to other
func (si *SyncInfo) Equal(other *SyncInfo) bool {
	if si == nil || other == nil {
		return si == other
	}

	if !si.State.Equal(other.State) {
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
