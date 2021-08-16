package ghostmanager

import "github.com/kaspanet/kaspad/domain/consensus/model"

type ghostManager struct {
	databaseContext model.DBManager

	dagTraversalManager model.DAGTraversalManager
	dagTopologyManager  model.DAGTopologyManager

	consensusStateStore model.ConsensusStateStore
}

// New instantiates a new GHOSTManager
func New(
	databaseContext model.DBManager,

	dagTraversalManager model.DAGTraversalManager,
	dagTopologyManager model.DAGTopologyManager,

	consensusStateStore model.ConsensusStateStore) model.GHOSTManager {

	return &ghostManager{
		databaseContext: databaseContext,

		dagTraversalManager: dagTraversalManager,
		dagTopologyManager:  dagTopologyManager,

		consensusStateStore: consensusStateStore,
	}
}
