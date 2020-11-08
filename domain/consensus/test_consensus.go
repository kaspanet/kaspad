package consensus

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type testConsensus struct{ *consensus }

func (tc *testConsensus) BuildBlockWithParents(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

	return tc.blockBuilder.BuildBlockWithParents(parentHashes, coinbaseData, transactions)
}
