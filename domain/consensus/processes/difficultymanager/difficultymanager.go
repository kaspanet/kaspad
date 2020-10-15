package difficultymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// DifficultyManager provides a method to resolve the
// difficulty value of a block
type difficultyManager struct {
	ghostdagManager model.GHOSTDAGManager
}

// New instantiates a new difficultyManager
func New(ghostdagManager model.GHOSTDAGManager) model.DifficultyManager {
	return &difficultyManager{
		ghostdagManager: ghostdagManager,
	}
}

// RequiredDifficulty returns the difficulty required for some block
func (d difficultyManager) RequiredDifficulty(parents []*model.DomainHash) uint32 {
	return 0
}
