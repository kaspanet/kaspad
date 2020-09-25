package state

import (
	"github.com/kaspanet/kaspad/domain/state/algorithms/blockprocessor/implementation"
	"github.com/kaspanet/kaspad/domain/state/algorithms/consensusstatemanager/implementation"
)

type Factory interface {
	NewState() State
}

type factory struct {
}

func (f *factory) NewState() State {
	blockProcessor := blockprocessor.New()
	consensusStateManager := consensusstatemanager.New()

	return &state{
		blockProcessor:        blockProcessor,
		consensusStateManager: consensusStateManager,
	}
}

func NewFactory() Factory {
	return &factory{}
}
