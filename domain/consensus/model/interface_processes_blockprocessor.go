package model

// BlockProcessor is responsible for processing incoming blocks
// and creating blocks from the current state
type BlockProcessor interface {
	BuildBlock(coinbaseScriptPublicKey []byte, coinbaseExtraData []byte, transactions []*DomainTransaction) (*DomainBlock, error)
	ValidateAndInsertBlock(block *DomainBlock) error
}
