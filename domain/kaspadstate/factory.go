package kaspadstate

import (
	"github.com/kaspanet/kaspad/domain/dagconfig"
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
	"github.com/kaspanet/kaspad/domain/kaspadstate/processes/blockprocessor"
	"github.com/kaspanet/kaspad/domain/kaspadstate/processes/blockvalidator"
	"github.com/kaspanet/kaspad/domain/kaspadstate/processes/consensusstatemanager"
	"github.com/kaspanet/kaspad/domain/kaspadstate/processes/dagtopologymanager"
	"github.com/kaspanet/kaspad/domain/kaspadstate/processes/dagtraversalmanager"
	"github.com/kaspanet/kaspad/domain/kaspadstate/processes/ghostdagmanager"
	"github.com/kaspanet/kaspad/domain/kaspadstate/processes/pruningmanager"
	"github.com/kaspanet/kaspad/domain/kaspadstate/processes/reachabilitytree"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

// Factory instantiates new KaspadStates
type Factory interface {
	NewKaspadState(dagParams *dagconfig.Params, databaseContext *dbaccess.DatabaseContext) KaspadState
}

type factory struct{}

// NewKaspadState instantiates a new KaspadState
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
	blockValidator := blockvalidator.New()
	reachabilityTree := reachabilitytree.New(
		blockRelationStore,
		reachabilityDataStore)
	dagTopologyManager := dagtopologymanager.New(
		reachabilityTree,
		blockRelationStore)
	ghostdagManager := ghostdagmanager.New(
		dagTopologyManager,
		ghostdagDataStore)
	dagTraversalManager := dagtraversalmanager.New(
		dagTopologyManager,
		ghostdagManager)
	pruningManager := pruningmanager.New(
		dagTraversalManager,
		pruningPointStore)
	consensusStateManager := consensusstatemanager.New(
		dagParams,
		consensusStateStore,
		multisetStore,
		utxoDiffStore)
	blockProcessor := blockprocessor.New(
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

// NewFactory creates a new KaspadState factory
func NewFactory() Factory {
	return &factory{}
}
