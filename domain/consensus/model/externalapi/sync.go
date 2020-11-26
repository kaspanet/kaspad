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

// SyncInfo holds info about the current sync state of the consensus
type SyncInfo struct {
	State                SyncState
	IBDRootUTXOBlockHash *DomainHash
	HeaderCount          uint64
	BlockCount           uint64
}

// Clone returns a clone of SyncInfo
func (si *SyncInfo) Clone() *SyncInfo {
	if si == nil {
		return nil
	}

	return &SyncInfo{
		State:                si.State.Clone(),
		IBDRootUTXOBlockHash: si.IBDRootUTXOBlockHash.Clone(),
		HeaderCount:          si.HeaderCount,
		BlockCount:           si.BlockCount,
	}
}
