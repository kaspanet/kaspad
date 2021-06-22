package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockProcessor is responsible for processing incoming blocks
type BlockProcessor interface {
	ValidateAndInsertBlock(block *externalapi.DomainBlock, validateUTXO bool) (*externalapi.BlockInsertionResult, error)
	ValidateAndInsertImportedPruningPoint(newPruningPoint *externalapi.DomainHash) error
	ValidateAndInsertBlockWithMetaData(block *externalapi.BlockWithMetaData, validateUTXO bool) (*externalapi.BlockInsertionResult, error)
}
