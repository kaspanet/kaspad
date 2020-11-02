package miningmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/miningmanager/blocktemplatebuilder"
	mempoolpkg "github.com/kaspanet/kaspad/domain/miningmanager/mempool"
)

// Factory instantiates new mining managers
type Factory interface {
	NewMiningManager(consensus consensus.Consensus, blockMaxMass uint64) MiningManager
}

type factory struct{}

// NewMiningManager instantiate a new mining manager
func (f *factory) NewMiningManager(consensus consensus.Consensus, blockMaxMass uint64) MiningManager {
	mempool := mempoolpkg.New(consensus)
	blockTemplateBuilder := blocktemplatebuilder.New(consensus, mempool, blockMaxMass)

	return &miningManager{
		mempool:              mempool,
		blockTemplateBuilder: blockTemplateBuilder,
	}
}

// NewFactory creates a new mining manager factory
func NewFactory() Factory {
	return &factory{}
}
