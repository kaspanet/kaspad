package blocktemplatebuilderimpl

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/kaspadstate"
)

type BlockTemplateBuilder struct {
	kaspadState *kaspadstate.KaspadState
}

func New(kaspadState *kaspadstate.KaspadState) *BlockTemplateBuilder {
	return &BlockTemplateBuilder{
		kaspadState: kaspadState,
	}
}

func (btb *BlockTemplateBuilder) GetBlockTemplate() *appmessage.MsgBlock {
	return nil
}
