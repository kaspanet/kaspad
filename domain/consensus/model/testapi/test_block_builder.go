package testapi

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// TestBlockBuilder adds to the main BlockBuilder methods required by tests
type TestBlockBuilder interface {
	model.BlockBuilder

	// BuildBlockWithParents builds a block with provided parents, coinbaseData and transactions,
	// and returns the block together with its past UTXO-diff from the virtual.
	BuildBlockWithParents(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
		transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, *model.UTXODiff, error)
}
