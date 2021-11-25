package parentssanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
)

type parentsManager struct {
	genesisHash *externalapi.DomainHash
}

// New instantiates a new ParentsManager
func New(genesisHash *externalapi.DomainHash) model.ParentsManager {
	return &parentsManager{
		genesisHash: genesisHash,
	}
}

func (pm *parentsManager) ParentsAtLevel(blockHeader externalapi.BlockHeader, level int) externalapi.BlockLevelParents {
	var parentsAtLevel externalapi.BlockLevelParents
	if len(blockHeader.Parents()) > level {
		parentsAtLevel = blockHeader.Parents()[level]
	}

	if len(parentsAtLevel) == 0 && len(blockHeader.DirectParents()) > 0 {
		return externalapi.BlockLevelParents{pm.genesisHash}
	}

	return parentsAtLevel
}

func (pm *parentsManager) Parents(blockHeader externalapi.BlockHeader) []externalapi.BlockLevelParents {
	numParents := constants.MaxBlockLevel + 1
	parents := make([]externalapi.BlockLevelParents, numParents)
	for i := 0; i < numParents; i++ {
		parents[i] = pm.ParentsAtLevel(blockHeader, i)
	}

	return parents
}
