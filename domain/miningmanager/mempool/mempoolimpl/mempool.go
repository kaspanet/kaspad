package mempoolimpl

import "github.com/kaspanet/kaspad/domain/state"

type Mempool struct {
	state *state.State
}

func New(state *state.State) *Mempool {
	return &Mempool{
		state: state,
	}
}
