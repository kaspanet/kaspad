package blocktemplatebuilderimpl

import "github.com/kaspanet/kaspad/domain/state"

type BlockTemplateBuilder struct {
	state *state.State
}

func New(state *state.State) *BlockTemplateBuilder {
	return &BlockTemplateBuilder{
		state: state,
	}
}
