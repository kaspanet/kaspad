package model

import "github.com/kaspanet/kaspad/util/daghash"

type BlockGHOSTDAGData struct {
}

func (bgd *BlockGHOSTDAGData) Blues() []*daghash.Hash {
	return nil
}

func (bgd *BlockGHOSTDAGData) BlueScore() uint64 {
	return 0
}

func (bgd *BlockGHOSTDAGData) Reds() []*daghash.Hash {
	return nil
}

func (bgd *BlockGHOSTDAGData) SelectedParent() *daghash.Hash {
	return nil
}
