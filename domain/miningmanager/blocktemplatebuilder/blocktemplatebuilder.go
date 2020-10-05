package blocktemplatebuilder

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/util"
)

// BlockTemplateBuilder ...
type BlockTemplateBuilder struct {
	consensus *consensus.Consensus
}

// New creates a new BlockTemplateBuilder
func New(consensus *consensus.Consensus) *BlockTemplateBuilder {
	return &BlockTemplateBuilder{
		consensus: consensus,
	}
}

// GetBlockTemplate creates a block template for a miner to consume
func (btb *BlockTemplateBuilder) GetBlockTemplate(payAddress util.Address, extraData []byte) *appmessage.MsgBlock {
	return nil
}
