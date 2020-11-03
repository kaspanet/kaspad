package externalapi

// Each of the following represent one of the possible sync
// states of the consensus
const (
	SyncStateRegular SyncState = iota
	SyncStateHeadersFirst
	SyncStateHeadersFirstAfterUTXOSet
	SyncStateIBD
)

// SyncState represents the current sync state of the consensus
type SyncState uint8

// SyncInfo holds info about the current sync state of the consensus
type SyncInfo struct {
	State                SyncState
	IBDRootUTXOBlockHash *DomainHash
}
