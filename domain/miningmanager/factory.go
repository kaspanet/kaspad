package miningmanager

import "github.com/kaspanet/kaspad/domain/state"

type Factory interface {
	NewMiningManager(state *state.State) MiningManager
}

type factory struct{}

func (f *factory) NewMiningManager(state *state.State) MiningManager {
	return &miningManager{
		mempool:              nil,
		blockTemplateBuilder: nil,
	}
}

func NewFactory() Factory {
	return &factory{}
}
