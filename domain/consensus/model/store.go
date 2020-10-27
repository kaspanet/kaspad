package model

type Store interface {
	Discard()
	Commit(dbTx DBTxProxy) error
}
