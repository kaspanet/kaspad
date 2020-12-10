package externalapi

// SyncInfo holds info about the current sync state of the consensus
type SyncInfo struct {
	IsAwaitingUTXOSet    bool
	IBDRootUTXOBlockHash *DomainHash
	HeaderCount          uint64
	BlockCount           uint64
}
