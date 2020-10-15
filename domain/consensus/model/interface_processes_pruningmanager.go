package model

// PruningManager resolves and manages the current pruning point
type PruningManager interface {
	FindNextPruningPoint(blockGHOSTDAGData *BlockGHOSTDAGData) (found bool,
		newPruningPoint *DomainHash, newPruningPointUTXOSet ReadOnlyUTXOSet, err error)
	PruningPoint() (*DomainHash, error)
	SerializedUTXOSet() ([]byte, error)
}
