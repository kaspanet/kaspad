package model

// BlockGHOSTDAGData represents GHOSTDAG data for some block
type BlockGHOSTDAGData struct {
	BlueScore          uint64
	SelectedParent     *DomainHash
	MergeSetBlues      []*DomainHash
	MergeSetReds       []*DomainHash
	BluesAnticoneSizes map[DomainHash]KType
}

// KType defines the size of GHOSTDAG consensus algorithm K parameter.
type KType byte
