package externalapi

const (
	SyncStateRegular SyncState = iota

	SyncStateHeadersFirst

	SyncStateHeadersFirstAfterUTXOSet

	SyncStateIBD
)

type SyncState uint8

type SyncInfo struct {
	State                SyncState
	IBDRootUTXOBlockHash *DomainHash
}
