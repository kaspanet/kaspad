package model

// BlockGHOSTDAGData represents GHOSTDAG data for some block
type BlockGHOSTDAGData struct {
	BlueScore      uint64
	SelectedParent *DomainHash
	MergeSetBlues  []*DomainHash
	MergeSetReds   []*DomainHash
}
