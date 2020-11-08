package testapi

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// TestConsensus wraps the Consensus interface with some methods that are needed by tests only
type TestConsensus interface {
	externalapi.Consensus

	BuildBlockWithParents(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
		transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error)
}
