package blockdag

import (
	"github.com/kaspanet/kaspad/infrastructure/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// IndexManager provides a generic interface that is called when blocks are
// connected to the DAG for the purpose of supporting optional indexes.
type IndexManager interface {
	// Init is invoked during DAG initialize in order to allow the index
	// manager to initialize itself and any indexes it is managing.
	Init(*BlockDAG, *dbaccess.DatabaseContext) error

	// ConnectBlock is invoked when a new block has been connected to the
	// DAG.
	ConnectBlock(dbContext *dbaccess.TxContext, blockHash *daghash.Hash, acceptedTxsData MultiBlockTxsAcceptanceData) error
}
