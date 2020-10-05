package model

import "github.com/kaspanet/kaspad/util/daghash"

// BlockGHOSTDAGData ...
type BlockGHOSTDAGData struct {
}

// MergeSetBlues ...
func (bgd *BlockGHOSTDAGData) MergeSetBlues() []*daghash.Hash {
	return nil
}

// BlueScore ...
func (bgd *BlockGHOSTDAGData) BlueScore() uint64 {
	return 0
}

// MergeSetReds ...
func (bgd *BlockGHOSTDAGData) MergeSetReds() []*daghash.Hash {
	return nil
}

// MergeSet ...
func (bgd *BlockGHOSTDAGData) MergeSet() []*daghash.Hash {
	return nil
}

// SelectedParent ...
func (bgd *BlockGHOSTDAGData) SelectedParent() *daghash.Hash {
	return nil
}
