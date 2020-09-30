package model

import "github.com/kaspanet/kaspad/util/daghash"

// BlockGHOSTDAGData ...
type BlockGHOSTDAGData struct {
}

// Blues ...
func (bgd *BlockGHOSTDAGData) Blues() []*daghash.Hash {
	return nil
}

// BlueScore ...
func (bgd *BlockGHOSTDAGData) BlueScore() uint64 {
	return 0
}

// Reds ...
func (bgd *BlockGHOSTDAGData) Reds() []*daghash.Hash {
	return nil
}

// SelectedParent ...
func (bgd *BlockGHOSTDAGData) SelectedParent() *daghash.Hash {
	return nil
}
