package blocktemplatebuilder

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	consensusexternalapi "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	miningmanagerapi "github.com/kaspanet/kaspad/domain/miningmanager/model"
)

// blockTemplateBuilder creates block templates for a miner to consume
type blockTemplateBuilder struct {
	consensus *consensus.Consensus
	mempool   miningmanagerapi.Mempool
}

// New creates a new blockTemplateBuilder
func New(consensus *consensus.Consensus, mempool miningmanagerapi.Mempool) miningmanagerapi.BlockTemplateBuilder {
	return &blockTemplateBuilder{
		consensus: consensus,
		mempool:   mempool,
	}
}

// GetBlockTemplate creates a block template for a miner to consume
func (btb *blockTemplateBuilder) GetBlockTemplate(coinbaseData *consensusexternalapi.DomainCoinbaseData) *consensusexternalapi.DomainBlock {
	return nil
}
