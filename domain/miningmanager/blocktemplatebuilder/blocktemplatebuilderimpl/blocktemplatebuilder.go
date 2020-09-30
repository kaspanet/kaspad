package blocktemplatebuilderimpl

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/kaspadstate"
)

// BlockTemplateBuilder ...
type BlockTemplateBuilder struct {
	kaspadState *kaspadstate.KaspadState
}

// New ...
func New(kaspadState *kaspadstate.KaspadState) *BlockTemplateBuilder {
	return &BlockTemplateBuilder{
		kaspadState: kaspadState,
	}
}

// GetBlockTemplate ...
func (btb *BlockTemplateBuilder) GetBlockTemplate() *appmessage.MsgBlock {
	return nil
}
