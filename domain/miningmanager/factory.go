package miningmanager

import (
	"github.com/kaspanet/kaspad/domain/miningmanager/blocktemplatebuilder/blocktemplatebuilderimpl"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/mempoolimpl"
	"github.com/kaspanet/kaspad/domain/state"
)

type Factory interface {
	NewMiningManager(state *state.State) MiningManager
}

type factory struct{}

func (f *factory) NewMiningManager(state *state.State) MiningManager {
	mempool := mempoolimpl.New(state)
	blockTemplateBuilder := blocktemplatebuilderimpl.New(state)

	return &miningManager{
		mempool:              mempool,
		blockTemplateBuilder: blockTemplateBuilder,
	}
}

func NewFactory() Factory {
	return &factory{}
}
