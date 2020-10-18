package blocktemplatebuilder

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	consensusmodel "github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/miningmanager/model"
)

// blockTemplateBuilder creates block templates for a miner to consume
type blockTemplateBuilder struct {
	consensus *consensus.Consensus
	mempool   model.Mempool
}

// New creates a new blockTemplateBuilder
func New(consensus *consensus.Consensus, mempool model.Mempool) model.BlockTemplateBuilder {
	return &blockTemplateBuilder{
		consensus: consensus,
		mempool:   mempool,
	}
}

// GetBlockTemplate creates a block template for a miner to consume
func (btb *blockTemplateBuilder) GetBlockTemplate(payAddress model.DomainAddress, extraData []byte) *consensusmodel.DomainBlock {
	return nil
}
