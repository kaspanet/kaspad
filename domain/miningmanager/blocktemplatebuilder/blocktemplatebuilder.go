package blocktemplatebuilder

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/kaspadstate"
	"github.com/kaspanet/kaspad/util"
)

// BlockTemplateBuilder ...
type BlockTemplateBuilder struct {
	kaspadState *kaspadstate.KaspadState
}

// New creates a new BlockTemplateBuilder
func New(kaspadState *kaspadstate.KaspadState) *BlockTemplateBuilder {
	return &BlockTemplateBuilder{
		kaspadState: kaspadState,
	}
}

// GetBlockTemplate creates a block template for a miner to consume
func (btb *BlockTemplateBuilder) GetBlockTemplate(payAddress util.Address, extraData []byte) *appmessage.MsgBlock {
	return nil
}
