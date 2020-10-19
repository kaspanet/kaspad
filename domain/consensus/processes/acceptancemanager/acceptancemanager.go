package acceptancemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// AcceptanceManager manages transaction acceptance
// and related data
type acceptanceManager struct {
	utxoDiffManager model.UTXODiffManager
}

// New instantiates a new AcceptanceManager
func New(utxoDiffManager model.UTXODiffManager) model.AcceptanceManager {
	return &acceptanceManager{
		utxoDiffManager: utxoDiffManager,
	}
}

func (a *acceptanceManager) CalculateAcceptanceDataAndUTXOMultiset(blockGHOSTDAGData *model.BlockGHOSTDAGData) (
	*model.BlockAcceptanceData, model.Multiset, error) {

	panic("implement me")
}
