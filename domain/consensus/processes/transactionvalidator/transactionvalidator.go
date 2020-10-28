package transactionvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// transactionValidator exposes a set of validation classes, after which
// it's possible to determine whether either a transaction is valid
type transactionValidator struct {
	blockCoinbaseMaturity uint64
	databaseContext       model.DBReader
	pastMedianTimeManager model.PastMedianTimeManager
	ghostdagDataStore     model.GHOSTDAGDataStore
}

// New instantiates a new TransactionValidator
func New(blockCoinbaseMaturity uint64,
	databaseContext model.DBReader,
	pastMedianTimeManager model.PastMedianTimeManager,
	ghostdagDataStore model.GHOSTDAGDataStore) model.TransactionValidator {
	return &transactionValidator{blockCoinbaseMaturity: blockCoinbaseMaturity,
		databaseContext:       databaseContext,
		pastMedianTimeManager: pastMedianTimeManager,
		ghostdagDataStore:     ghostdagDataStore,
	}
}
