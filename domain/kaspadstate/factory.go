package kaspadstate

import (
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/blockprocessor/blockprocessorimpl"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/blockvalidator/blockvalidatorimpl"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/consensusstatemanager/consensusstatemanagerimpl"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/dagtopologymanager/dagtopologymanagerimpl"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/dagtraversalmanager/dagtraversalmanagerimpl"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/ghostdagmanager/ghostdagmanagerimpl"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/pruningmanager/pruningmanagerimpl"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/reachabilitytree/reachabilitytreeimpl"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/acceptancedatastore/acceptancedatastoreimpl"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/blockindex/blockindeximpl"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/blockmessagestore/blockmessagestoreimpl"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/blockrelationstore/blockrelationstoreimpl"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/blockstatusstore/blockstatusstoreimpl"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/consensusstatestore/consensusstatestoreimpl"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/ghostdagdatastore/ghostdagdatastoreimpl"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/multisetstore/multisetstoreimpl"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/pruningpointstore/pruningpointstoreimpl"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/reachabilitydatastore/reachabilitydatastoreimpl"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/utxodiffstore/utxodiffstoreimpl"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

type Factory interface {
	NewKaspadState(dagParams *dagconfig.Params, databaseContext *dbaccess.DatabaseContext) KaspadState
}

type factory struct{}

func (f *factory) NewKaspadState(dagParams *dagconfig.Params, databaseContext *dbaccess.DatabaseContext) KaspadState {
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

	return &kaspadState{
		consensusStateManager: consensusStateManager,
		blockProcessor:        blockProcessor,
	}
}

func NewFactory() Factory {
	return &factory{}
}
