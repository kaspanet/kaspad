package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockProcessor is responsible for processing incoming blocks
// and creating blocks from the current state
type BlockProcessor interface {
	BuildBlock(coinbaseScriptPublicKey []byte, coinbaseExtraData []byte, transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error)
	ValidateAndInsertBlock(block *externalapi.DomainBlock) error
}
