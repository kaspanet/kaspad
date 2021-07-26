package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockProcessor is responsible for processing incoming blocks
type BlockProcessor interface {
	ValidateAndInsertBlock(block *externalapi.DomainBlock, shouldValidateAgainstUTXO bool) (*externalapi.BlockInsertionResult, error)
	ValidateAndInsertImportedPruningPoint(newPruningPoint *externalapi.DomainHash) error
	ValidateAndInsertBlockWithTrustedData(block *externalapi.BlockWithTrustedData, validateUTXO bool) (*externalapi.BlockInsertionResult, error)
}
