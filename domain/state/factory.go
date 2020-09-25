package state

import (
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/state/algorithms/blockprocessor/blockprocessorimpl"
	"github.com/kaspanet/kaspad/domain/state/algorithms/consensusstatemanager/consensusstatemanagerimpl"
	"github.com/kaspanet/kaspad/domain/state/datastructures/consensusstatestore/consensusstatestoreimpl"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

type Factory interface {
	NewState(dagParams *dagconfig.Params, databaseContext *dbaccess.DatabaseContext) State
}

type factory struct {
}

func (f *factory) NewState(dagParams *dagconfig.Params, databaseContext *dbaccess.DatabaseContext) State {
	consensusStateStore := consensusstatestoreimpl.New()

	consensusStateManager := consensusstatemanagerimpl.New(dagParams, consensusStateStore)
	blockProcessor := blockprocessorimpl.New(dagParams, databaseContext, consensusStateManager)

	return &state{
		consensusStateManager: consensusStateManager,
		blockProcessor:        blockProcessor,
	}
}

func NewFactory() Factory {
	return &factory{}
}
