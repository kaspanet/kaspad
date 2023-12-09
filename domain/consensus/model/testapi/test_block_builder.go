package testapi

import (
	"github.com/zoomy-network/zoomyd/domain/consensus/model"
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
)

// TestBlockBuilder adds to the main BlockBuilder methods required by tests
type TestBlockBuilder interface {
	model.BlockBuilder

	// BuildBlockWithParents builds a block with provided parents, coinbaseData and transactions,
	// and returns the block together with its past UTXO-diff from the virtual.
	BuildBlockWithParents(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
		transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, externalapi.UTXODiff, error)

	BuildUTXOInvalidHeader(parentHashes []*externalapi.DomainHash) (externalapi.BlockHeader, error)

	BuildUTXOInvalidBlock(parentHashes []*externalapi.DomainHash) (*externalapi.DomainBlock,
		error)

	SetNonceCounter(nonceCounter uint64)
}
