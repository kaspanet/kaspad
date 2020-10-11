package model

// PruningManager resolves and manages the current pruning point
type PruningManager interface {
	FindNextPruningPoint(blockHash *DomainHash) (found bool, newPruningPoint *DomainHash, newPruningPointUTXOSet ReadOnlyUTXOSet)
	PruningPoint() *DomainHash
	SerializedUTXOSet() []byte
}
