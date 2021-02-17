package consensus

import "github.com/kaspanet/kaspad/domain/consensus/model"

// GHOSTDAGManagerConstructor is the function signature for a constructor of a type implementing model.GHOSTDAGManager
type GHOSTDAGManagerConstructor func(model.DBReader, model.DAGTopologyManager, model.GHOSTDAGDataStore, model.BlockHeaderStore, model.KType) model.GHOSTDAGManager
