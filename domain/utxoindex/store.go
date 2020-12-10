package utxoindex

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

type utxoIndexStore struct {
	database database.Database
}

func newUTXOIndexStore(database database.Database) *utxoIndexStore {
	return &utxoIndexStore{
		database: database,
	}
}

func (uis *utxoIndexStore) add(scriptPublicKey []byte, outpoint *externalapi.DomainOutpoint, utxoEntry *externalapi.UTXOEntry) {

}

func (uis *utxoIndexStore) remove(scriptPublicKey []byte, outpoint *externalapi.DomainOutpoint) {

}

func (uis *utxoIndexStore) commit() error {
	return nil
}
