package model

// BlockGHOSTDAGData represents GHOSTDAG data for some block
type BlockGHOSTDAGData struct {
	blueScore      uint64
	selectedParent *DomainHash
	mergeSetBlues  []*DomainHash
	mergeSetReds   []*DomainHash
}

// MergeSetBlues returns the merge-set blues of this block
func (bgd *BlockGHOSTDAGData) MergeSetBlues() []*DomainHash {
	return nil
}

// BlueScore returns the blue score of this block
func (bgd *BlockGHOSTDAGData) BlueScore() uint64 {
	return 0
}

// MergeSetReds returns the merge-set reds of this block
func (bgd *BlockGHOSTDAGData) MergeSetReds() []*DomainHash {
	return nil
}

// MergeSet returns the entire merge-set of this block
func (bgd *BlockGHOSTDAGData) MergeSet() []*DomainHash {
	return nil
}

// SelectedParent returns this block's selected parent
func (bgd *BlockGHOSTDAGData) SelectedParent() *DomainHash {
	return nil
}
