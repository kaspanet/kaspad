package consensus

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type testConsensus struct {
	*consensus
	testBlockBuilder model.TestBlockBuilder
}

func (tc *testConsensus) BuildBlockWithParents(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

	return tc.testBlockBuilder.BuildBlockWithParents(parentHashes, coinbaseData, transactions)
}
