package difficultymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// DifficultyManager provides a method to resolve the
// difficulty value of a block
type difficultyManager struct {
	ghostdagManager model.GHOSTDAGManager
}

// New instantiates a new DifficultyManager
func New(ghostdagManager model.GHOSTDAGManager) model.DifficultyManager {
	return &difficultyManager{
		ghostdagManager: ghostdagManager,
	}
}

// RequiredDifficulty returns the difficulty required for some block
func (dm *difficultyManager) RequiredDifficulty(blockHash *externalapi.DomainHash) (uint32, error) {
	return 0, nil
}
