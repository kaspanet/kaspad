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
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/acceptancedatastore"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/blockindex"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/blockmessagestore"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/blockrelationstore"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/blockstatusstore"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/consensusstatestore"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/ghostdagdatastore"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/multisetstore"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/pruningpointstore"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/reachabilitydatastore"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/utxodiffstore"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

type Factory interface {
	NewKaspadState(dagParams *dagconfig.Params, databaseContext *dbaccess.DatabaseContext) KaspadState
}

type factory struct{}

func (f *factory) NewKaspadState(dagParams *dagconfig.Params, databaseContext *dbaccess.DatabaseContext) KaspadState {
	// Data Structures
	acceptanceDataStore := acceptancedatastore.New()
	blockIndex := blockindex.New()
	blockMessageStore := blockmessagestore.New()
	blockRelationStore := blockrelationstore.New()
	blockStatusStore := blockstatusstore.New()
	multisetStore := multisetstore.New()
	pruningPointStore := pruningpointstore.New()
	reachabilityDataStore := reachabilitydatastore.New()
	utxoDiffStore := utxodiffstore.New()
	consensusStateStore := consensusstatestore.New()
	ghostdagDataStore := ghostdagdatastore.New()

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
