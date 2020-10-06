package miningmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/miningmanager/blocktemplatebuilder"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool"
)

// Factory instantiates new mining managers
type Factory interface {
	NewMiningManager(consensus *consensus.Consensus) MiningManager
}

type factory struct{}

// NewMiningManager instantiate a new mining manager
func (f *factory) NewMiningManager(consensus *consensus.Consensus) MiningManager {
	return &miningManager{
		mempool:              mempool.New(consensus),
		blockTemplateBuilder: blocktemplatebuilder.New(consensus),
	}
}

// NewFactory creates a new mining manager factory
func NewFactory() Factory {
	return &factory{}
}
