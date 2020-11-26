package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockBuilder is responsible for creating blocks from the current state
type BlockBuilder interface {
	BuildBlock(coinbaseData *externalapi.DomainCoinbaseData, transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error)
}

// TestBlockBuilder adds to the main BlockBuilder methods required by tests
type TestBlockBuilder interface {
	BlockBuilder

	// BuildBlockWithParents builds a block with provided parents, coinbaseData and transactions,
	// and returns the block together with its past UTXO-diff.
	BuildBlockWithParents(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
		transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, *UTXODiff, error)
}
