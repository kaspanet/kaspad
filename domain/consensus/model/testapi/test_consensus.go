package testapi

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type TestConsensus interface {
	externalapi.Consensus

	BuildBlockWithParents(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
		transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error)
}
