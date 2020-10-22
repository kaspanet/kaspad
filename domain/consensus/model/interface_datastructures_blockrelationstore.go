package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockRelationStore represents a store of BlockRelations
type BlockRelationStore interface {
	Stage(blockHash *externalapi.DomainHash, parentHashes []*externalapi.DomainHash)
	IsStaged() bool
	Discard()
	Commit(dbTx DBTxProxy) error
	Get(dbContext DBContextProxy, blockHash *externalapi.DomainHash) (*BlockRelations, error)
}
