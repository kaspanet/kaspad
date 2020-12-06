package externalapi

import "fmt"

// Each of the following represent one of the possible sync
// states of the consensus
const (
	SyncStateSynced SyncState = iota
	SyncStateAwaitingGenesis
	SyncStateAwaitingUTXOSet
	SyncStateAwaitingBlockBodies
)

// SyncState represents the current sync state of the consensus
type SyncState uint8

func (s SyncState) String() string {
	switch s {
	case SyncStateSynced:
		return "SyncStateSynced"
	case SyncStateAwaitingGenesis:
		return "SyncStateAwaitingGenesis"
	case SyncStateAwaitingUTXOSet:
		return "SyncStateAwaitingUTXOSet"
	case SyncStateAwaitingBlockBodies:
		return "SyncStateAwaitingBlockBodies"
	}

	return fmt.Sprintf("<unknown state (%d)>", s)
}

// SyncInfo holds info about the current sync state of the consensus
type SyncInfo struct {
	State                SyncState
	IBDRootUTXOBlockHash *DomainHash
	HeaderCount          uint64
	BlockCount           uint64
}
