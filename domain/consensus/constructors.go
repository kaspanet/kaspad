package consensus

import "github.com/kaspanet/kaspad/domain/consensus/model"

// GHOSTDAGManagerConstructor is the function signature for a constructor of a type implementing model.GHOSTDAGManager
type GHOSTDAGManagerConstructor func(model.DBReader, model.DAGTopologyManager, model.GHOSTDAGDataStore, model.BlockHeaderStore, model.KType) model.GHOSTDAGManager

// MEDIAN is the function signature for a constructor of a type implementing model.PastMedianTimeManager
type PastMedianTimeManagerConstructor func(int, model.DBReader, model.DAGTraversalManager, model.BlockHeaderStore, model.GHOSTDAGDataStore) model.PastMedianTimeManager
