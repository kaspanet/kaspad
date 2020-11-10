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
	case SyncStateHeadersFirst:
		return "SyncStateHeadersFirst"
	case SyncStateMissingUTXOSet:
		return "SyncStateMissingUTXOSet"
	case SyncStateMissingBlockBodies:
		return "SyncStateMissingBlockBodies"
	}

	return fmt.Sprintf("<unknown state (%d)>", s)
}

// SyncInfo holds info about the current sync state of the consensus
type SyncInfo struct {
	State                SyncState
	IBDRootUTXOBlockHash *DomainHash
}
