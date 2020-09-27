package state

import (
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/state/algorithms/blockprocessor/blockprocessorimpl"
	"github.com/kaspanet/kaspad/domain/state/algorithms/blockvalidator/blockvalidatorimpl"
	"github.com/kaspanet/kaspad/domain/state/algorithms/consensusstatemanager/consensusstatemanagerimpl"
	"github.com/kaspanet/kaspad/domain/state/algorithms/dagtopologymanager/dagtopologymanagerimpl"
	"github.com/kaspanet/kaspad/domain/state/algorithms/dagtraversalmanager/dagtraversalmanagerimpl"
	"github.com/kaspanet/kaspad/domain/state/algorithms/ghostdagmanager/ghostdagmanagerimpl"
	"github.com/kaspanet/kaspad/domain/state/algorithms/pruningmanager/pruningmanagerimpl"
	"github.com/kaspanet/kaspad/domain/state/algorithms/reachabilitytree/reachabilitytreeimpl"
	"github.com/kaspanet/kaspad/domain/state/datastructures/acceptancedatastore/acceptancedatastoreimpl"
	"github.com/kaspanet/kaspad/domain/state/datastructures/blockindex/blockindeximpl"
	"github.com/kaspanet/kaspad/domain/state/datastructures/blockmessagestore/blockmessagestoreimpl"
	"github.com/kaspanet/kaspad/domain/state/datastructures/blockrelationstore/blockrelationstoreimpl"
	"github.com/kaspanet/kaspad/domain/state/datastructures/blockstatusstore/blockstatusstoreimpl"
	"github.com/kaspanet/kaspad/domain/state/datastructures/consensusstatestore/consensusstatestoreimpl"
	"github.com/kaspanet/kaspad/domain/state/datastructures/ghostdagdatastore/ghostdagdatastoreimpl"
	"github.com/kaspanet/kaspad/domain/state/datastructures/multisetstore/multisetstoreimpl"
	"github.com/kaspanet/kaspad/domain/state/datastructures/pruningpointstore/pruningpointstoreimpl"
	"github.com/kaspanet/kaspad/domain/state/datastructures/reachabilitydatastore/reachabilitydatastoreimpl"
	"github.com/kaspanet/kaspad/domain/state/datastructures/utxodiffstore/utxodiffstoreimpl"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

type Factory interface {
	NewState(dagParams *dagconfig.Params, databaseContext *dbaccess.DatabaseContext) State
}

type factory struct{}

func (f *factory) NewState(dagParams *dagconfig.Params, databaseContext *dbaccess.DatabaseContext) State {
	// Data Structures
	acceptanceDataStore := acceptancedatastoreimpl.New()
	blockIndex := blockindeximpl.New()
	blockMessageStore := blockmessagestoreimpl.New()
	blockRelationStore := blockrelationstoreimpl.New()
	blockStatusStore := blockstatusstoreimpl.New()
	multisetStore := multisetstoreimpl.New()
	pruningPointStore := pruningpointstoreimpl.New()
	reachabilityDataStore := reachabilitydatastoreimpl.New()
	utxoDiffStore := utxodiffstoreimpl.New()
	consensusStateStore := consensusstatestoreimpl.New()
	ghostdagDataStore := ghostdagdatastoreimpl.New()

	// Algorithms
	blockValidator := blockvalidatorimpl.New()
	reachabilityTree := reachabilitytreeimpl.New(
		blockRelationStore,
		reachabilityDataStore)
	dagTopologyManager := dagtopologymanagerimpl.New(
		reachabilityTree,
		blockRelationStore)
	ghostdagManager := ghostdagmanagerimpl.New(
		dagTopologyManager,
		ghostdagDataStore)
	dagTraversalManager := dagtraversalmanagerimpl.New(
		dagTopologyManager,
		ghostdagManager)
	pruningManager := pruningmanagerimpl.New(
		dagTraversalManager,
		pruningPointStore)
	consensusStateManager := consensusstatemanagerimpl.New(
		dagParams,
		consensusStateStore,
		multisetStore,
		utxoDiffStore)
	blockProcessor := blockprocessorimpl.New(
		dagParams,
		databaseContext,
		consensusStateManager,
		pruningManager,
		blockValidator,
		dagTopologyManager,
		reachabilityTree,
		acceptanceDataStore,
		blockIndex,
		blockMessageStore,
		blockStatusStore)

	return &state{
		consensusStateManager: consensusStateManager,
		blockProcessor:        blockProcessor,
	}
}

func NewFactory() Factory {
	return &factory{}
}
