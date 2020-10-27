package model

// Store is a common interface for data stores
type Store interface {
	Discard()
	Commit(dbTx DBTxProxy) error
}
