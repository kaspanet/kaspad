package miningmanager

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate"
	"github.com/kaspanet/kaspad/domain/miningmanager/blocktemplatebuilder/blocktemplatebuilderimpl"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/mempoolimpl"
)

type Factory interface {
	NewMiningManager(kaspadState *kaspadstate.KaspadState) MiningManager
}

type factory struct{}

func (f *factory) NewMiningManager(kaspadState *kaspadstate.KaspadState) MiningManager {
	mempool := mempoolimpl.New(kaspadState)
	blockTemplateBuilder := blocktemplatebuilderimpl.New(kaspadState)

	return &miningManager{
		mempool:              mempool,
		blockTemplateBuilder: blockTemplateBuilder,
	}
}

func NewFactory() Factory {
	return &factory{}
}
