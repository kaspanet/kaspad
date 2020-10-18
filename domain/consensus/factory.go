package consensus

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/acceptancedatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockrelationstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockstatusstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/consensusstatestore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/feedatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/ghostdagdatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/multisetstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/pruningstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/reachabilitydatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/utxodiffstore"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/processes/blockprocessor"
	"github.com/kaspanet/kaspad/domain/consensus/processes/consensusstatemanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtopologymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtraversalmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/difficultymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdagmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/pastmediantimemanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/pruningmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/reachabilitytree"
	validatorpkg "github.com/kaspanet/kaspad/domain/consensus/processes/validator"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

// Factory instantiates new Consensuses
type Factory interface {
	NewConsensus(dagParams *dagconfig.Params, databaseContext *dbaccess.DatabaseContext) Consensus
}

type factory struct{}

// NewConsensus instantiates a new Consensus
func (f *factory) NewConsensus(dagParams *dagconfig.Params, databaseContext *dbaccess.DatabaseContext) Consensus {
	// Data Structures
	acceptanceDataStore := acceptancedatastore.New()
	blockStore := blockstore.New()
	blockRelationStore := blockrelationstore.New()
	blockStatusStore := blockstatusstore.New()
	multisetStore := multisetstore.New()
	pruningStore := pruningstore.New()
	reachabilityDataStore := reachabilitydatastore.New()
	utxoDiffStore := utxodiffstore.New()
	consensusStateStore := consensusstatestore.New()
	ghostdagDataStore := ghostdagdatastore.New()
	feeDataStore := feedatastore.New()

	domainDBContext := database.NewDomainDBContext(databaseContext)

	// Processes
	reachabilityTree := reachabilitytree.New(
		blockRelationStore,
		reachabilityDataStore)
	dagTopologyManager := dagtopologymanager.New(
		domainDBContext,
		reachabilityTree,
		blockRelationStore)
	ghostdagManager := ghostdagmanager.New(
		databaseContext,
		dagTopologyManager,
		ghostdagDataStore,
		model.KType(dagParams.K))
	dagTraversalManager := dagtraversalmanager.New(
		dagTopologyManager,
		ghostdagManager)
	consensusStateManager := consensusstatemanager.New(
		domainDBContext,
		dagParams,
		consensusStateStore,
		multisetStore,
		utxoDiffStore,
		blockStore,
		ghostdagManager)
	pruningManager := pruningmanager.New(
		dagTraversalManager,
		pruningStore,
		dagTopologyManager,
		blockStatusStore,
		consensusStateManager)
	difficultyManager := difficultymanager.New(
		ghostdagManager)
	pastMedianTimeManager := pastmediantimemanager.New(
		ghostdagManager)
	validator := validatorpkg.New(
		consensusStateManager,
		difficultyManager,
		pastMedianTimeManager)
	blockProcessor := blockprocessor.New(
		dagParams,
		domainDBContext,
		consensusStateManager,
		pruningManager,
		validator,
		dagTopologyManager,
		reachabilityTree,
		difficultyManager,
		pastMedianTimeManager,
		ghostdagManager,
		acceptanceDataStore,
		blockStore,
		blockStatusStore,
		feeDataStore)

	return &consensus{
		consensusStateManager: consensusStateManager,
		blockProcessor:        blockProcessor,
		transactionValidator:  validator,
	}
}

// NewFactory creates a new Consensus factory
func NewFactory() Factory {
	return &factory{}
}
