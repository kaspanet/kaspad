package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockRelationStore represents a store of BlockRelations
type BlockRelationStore interface {
	Update(dbTx DBTxProxy, blockHash *externalapi.DomainHash, parentHashes []*externalapi.DomainHash) error
	Get(dbContext DBContextProxy, blockHash *externalapi.DomainHash) (*BlockRelations, error)
}
