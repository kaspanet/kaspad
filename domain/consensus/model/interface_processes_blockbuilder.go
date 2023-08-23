package model

import "github.com/c4ei/kaspad/domain/consensus/model/externalapi"

// BlockBuilder is responsible for creating blocks from the current state
type BlockBuilder interface {
	BuildBlock(coinbaseData *externalapi.DomainCoinbaseData,
		transactions []*externalapi.DomainTransaction) (block *externalapi.DomainBlock, coinbaseHasRedReward bool, err error)
}
