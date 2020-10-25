package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// PruningManager resolves and manages the current pruning point
type PruningManager interface {
	FindNextPruningPoint(blockHash *externalapi.DomainHash) error
	PruningPoint() (*externalapi.DomainHash, error)
	SerializedUTXOSet() ([]byte, error)
}
