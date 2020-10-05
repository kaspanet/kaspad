package model

import "github.com/kaspanet/kaspad/util/daghash"

// BlockGHOSTDAGData represents GHOSTDAG data for some block
type BlockGHOSTDAGData struct {
	blueScore      uint64
	selectedParent *daghash.Hash
	mergeSetBlues  []*daghash.Hash
	mergeSetReds   []*daghash.Hash
}

// MergeSetBlues returns the merge-set blues of this block
func (bgd *BlockGHOSTDAGData) MergeSetBlues() []*daghash.Hash {
	return nil
}

// BlueScore returns the blue score of this block
func (bgd *BlockGHOSTDAGData) BlueScore() uint64 {
	return 0
}

// MergeSetReds returns the merge-set reds of this block
func (bgd *BlockGHOSTDAGData) MergeSetReds() []*daghash.Hash {
	return nil
}

// MergeSet returns the entire merge-set of this block
func (bgd *BlockGHOSTDAGData) MergeSet() []*daghash.Hash {
	return nil
}

// SelectedParent returns this block's selected parent
func (bgd *BlockGHOSTDAGData) SelectedParent() *daghash.Hash {
	return nil
}
