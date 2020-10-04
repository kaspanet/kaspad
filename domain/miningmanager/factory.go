package miningmanager

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate"
	"github.com/kaspanet/kaspad/domain/miningmanager/blocktemplatebuilder"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool"
)

// Factory instantiates new mining managers
type Factory interface {
	NewMiningManager(kaspadState *kaspadstate.KaspadState) MiningManager
}

type factory struct{}

// NewMiningManager instantiate a new mining manager
func (f *factory) NewMiningManager(kaspadState *kaspadstate.KaspadState) MiningManager {
	return &miningManager{
		mempool:              mempool.New(kaspadState),
		blockTemplateBuilder: blocktemplatebuilder.New(kaspadState),
	}
}

// NewFactory creates a new mining manager factory
func NewFactory() Factory {
	return &factory{}
}
