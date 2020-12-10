package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// PruningManager resolves and manages the current pruning point
type PruningManager interface {
	UpdatePruningPointByVirtual() error
	CalculateIndependentPruningPoint(blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error)
}
