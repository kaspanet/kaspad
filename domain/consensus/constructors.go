package consensus

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"math/big"
	"time"
)

// GHOSTDAGManagerConstructor is the function signature for a constructor of a type implementing model.GHOSTDAGManager
type GHOSTDAGManagerConstructor func(model.DBReader, model.DAGTopologyManager,
	model.GHOSTDAGDataStore, model.BlockHeaderStore, model.KType) model.GHOSTDAGManager

// DifficultyManagerConstructor is the function signature for a constructor of a type implementing model.DifficultyManager
type DifficultyManagerConstructor func(model.DBReader, model.GHOSTDAGManager, model.GHOSTDAGDataStore,
	model.BlockHeaderStore, model.DAGTopologyManager, model.DAGTraversalManager, *big.Int, int, bool, time.Duration,
	*externalapi.DomainHash) model.DifficultyManager

// PastMedianTimeManagerConstructor is the function signature for a constructor of a type implementing model.PastMedianTimeManager
type PastMedianTimeManagerConstructor func(int, model.DBReader, model.DAGTraversalManager, model.BlockHeaderStore,
	model.GHOSTDAGDataStore) model.PastMedianTimeManager
