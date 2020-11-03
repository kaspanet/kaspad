package consensus

import (
	consensusdatabase "github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/acceptancedatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockheaderstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockrelationstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockstatusstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/consensusstatestore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/ghostdagdatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/headertipsstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/multisetstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/pruningstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/reachabilitydatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/utxodiffstore"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/blockprocessor"
	"github.com/kaspanet/kaspad/domain/consensus/processes/blockvalidator"
	"github.com/kaspanet/kaspad/domain/consensus/processes/coinbasemanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/consensusstatemanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtopologymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtraversalmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/difficultymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdagmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/headertipsmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/mergedepthmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/pastmediantimemanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/pruningmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/reachabilitymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/syncmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/transactionvalidator"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	infrastructuredatabase "github.com/kaspanet/kaspad/infrastructure/db/database"
)

// Factory instantiates new Consensuses
type Factory interface {
	NewConsensus(dagParams *dagconfig.Params, db infrastructuredatabase.Database) (Consensus, error)
}

type factory struct{}

// NewConsensus instantiates a new Consensus
func (f *factory) NewConsensus(dagParams *dagconfig.Params, db infrastructuredatabase.Database) (Consensus, error) {
	// Data Structures
	acceptanceDataStore := acceptancedatastore.New()
	blockStore := blockstore.New()
	blockHeaderStore := blockheaderstore.New()
	blockRelationStore := blockrelationstore.New()
	blockStatusStore := blockstatusstore.New()
	multisetStore := multisetstore.New()
	pruningStore := pruningstore.New()
	reachabilityDataStore := reachabilitydatastore.New()
	utxoDiffStore := utxodiffstore.New()
	consensusStateStore := consensusstatestore.New()
	ghostdagDataStore := ghostdagdatastore.New()
	headerTipsStore := headertipsstore.New()

	dbManager := consensusdatabase.New(db)

	// Processes
	reachabilityManager := reachabilitymanager.New(
		dbManager,
		ghostdagDataStore,
		blockRelationStore,
		reachabilityDataStore)
	dagTopologyManager := dagtopologymanager.New(
		dbManager,
		reachabilityManager,
		blockRelationStore)
	ghostdagManager := ghostdagmanager.New(
		dbManager,
		dagTopologyManager,
		ghostdagDataStore,
		model.KType(dagParams.K))
	dagTraversalManager := dagtraversalmanager.New(
		dbManager,
		dagTopologyManager,
		ghostdagDataStore,
		ghostdagManager)
	pruningManager := pruningmanager.New(
		dagTraversalManager,
		dagTopologyManager,
		pruningStore,
		blockStatusStore,
		consensusStateStore)
	pastMedianTimeManager := pastmediantimemanager.New(
		dagParams.TimestampDeviationTolerance,
		dbManager,
		dagTraversalManager,
		blockHeaderStore)
	transactionValidator := transactionvalidator.New(dagParams.BlockCoinbaseMaturity,
		dbManager,
		pastMedianTimeManager,
		ghostdagDataStore)
	difficultyManager := difficultymanager.New(
		ghostdagManager)
	coinbaseManager := coinbasemanager.New(
		dbManager,
		ghostdagDataStore,
		acceptanceDataStore)
	headerTipsManager := headertipsmanager.New(dbManager, dagTopologyManager, headerTipsStore)
	genesisHash := externalapi.DomainHash(*dagParams.GenesisHash)
	mergeDepthManager := mergedepthmanager.New(
		dagParams.FinalityDepth(),
		dbManager,
		dagTopologyManager,
		dagTraversalManager,
		ghostdagDataStore)
	blockValidator := blockvalidator.New(
		dagParams.PowMax,
		false,
		&genesisHash,
		dagParams.EnableNonNativeSubnetworks,
		dagParams.DisableDifficultyAdjustment,
		dagParams.DifficultyAdjustmentWindowSize,

		dbManager,
		difficultyManager,
		pastMedianTimeManager,
		transactionValidator,
		ghostdagManager,
		dagTopologyManager,
		dagTraversalManager,
		coinbaseManager,
		mergeDepthManager,

		blockStore,
		ghostdagDataStore,
		blockHeaderStore,
		blockStatusStore,
	)
	consensusStateManager, err := consensusstatemanager.New(
		dbManager,
		dagParams,
		ghostdagManager,
		dagTopologyManager,
		dagTraversalManager,
		pruningManager,
		pastMedianTimeManager,
		transactionValidator,
		blockValidator,
		reachabilityManager,
		coinbaseManager,
		mergeDepthManager,

		blockStatusStore,
		ghostdagDataStore,
		consensusStateStore,
		multisetStore,
		blockStore,
		utxoDiffStore,
		blockRelationStore,
		acceptanceDataStore,
		blockHeaderStore,
		headerTipsStore)
	if err != nil {
		return nil, err
	}

	syncManager := syncmanager.New(
		dbManager,
		&genesisHash,
		dagTraversalManager,
		dagTopologyManager,
		ghostdagDataStore,
		blockStatusStore)

	blockProcessor := blockprocessor.New(
		dagParams,
		dbManager,
		consensusStateManager,
		pruningManager,
		blockValidator,
		dagTopologyManager,
		reachabilityManager,
		difficultyManager,
		pastMedianTimeManager,
		ghostdagManager,
		coinbaseManager,
		headerTipsManager,
		syncManager,

		acceptanceDataStore,
		blockStore,
		blockStatusStore,
		blockRelationStore,
		multisetStore,
		ghostdagDataStore,
		consensusStateStore,
		pruningStore,
		reachabilityDataStore,
		utxoDiffStore,
		blockHeaderStore,
		headerTipsStore)

	return &consensus{
		databaseContext: dbManager,

		blockProcessor:        blockProcessor,
		consensusStateManager: consensusStateManager,
		transactionValidator:  transactionValidator,
		syncManager:           syncManager,
		pastMedianTimeManager: pastMedianTimeManager,

		blockStore:        blockStore,
		blockHeaderStore:  blockHeaderStore,
		pruningStore:      pruningStore,
		ghostdagDataStore: ghostdagDataStore,
		blockStatusStore:  blockStatusStore,
	}, nil
}

// NewFactory creates a new Consensus factory
func NewFactory() Factory {
	return &factory{}
}
