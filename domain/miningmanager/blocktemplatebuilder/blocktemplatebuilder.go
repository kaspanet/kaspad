package blocktemplatebuilder

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	consensusmodel "github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/miningmanager/model"
	"github.com/kaspanet/kaspad/util"
)

// BlockTemplateBuilder creates block templates for a miner to consume
type BlockTemplateBuilder struct {
	consensus *consensus.Consensus
	mempool   model.Mempool
}

// New creates a new BlockTemplateBuilder
func New(consensus *consensus.Consensus, mempool model.Mempool) *BlockTemplateBuilder {
	return &BlockTemplateBuilder{
		consensus: consensus,
		mempool:   mempool,
	}
}

// GetBlockTemplate creates a block template for a miner to consume
func (btb *BlockTemplateBuilder) GetBlockTemplate(payAddress util.Address, extraData []byte) *consensusmodel.DomainBlock {
	return nil
}
