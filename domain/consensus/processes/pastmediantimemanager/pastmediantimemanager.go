package pastmediantimemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// pastMedianTimeManager provides a method to resolve the
// past median time of a block
type pastMedianTimeManager struct {
	ghostdagManager model.GHOSTDAGManager
}

// New instantiates a new PastMedianTimeManager
func New(ghostdagManager model.GHOSTDAGManager) model.PastMedianTimeManager {
	return &pastMedianTimeManager{
		ghostdagManager: ghostdagManager,
	}
}

// PastMedianTime returns the past median time for some block
func (pmtm *pastMedianTimeManager) PastMedianTime(blockHash *externalapi.DomainHash) (int64, error) {
	return 0, nil
}
