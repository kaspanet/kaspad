package externalapi

import "math/big"

// BlockInfo contains various information about a specific block
type BlockInfo struct {
	Exists         bool
	BlockStatus    BlockStatus
	BlueScore      uint64
	BlueWork       *big.Int
	SelectedParent *DomainHash
	MergeSetBlues  []*DomainHash
	MergeSetReds   []*DomainHash
}

// HasHeader returns whether the block exists and has a valid header
func (bi *BlockInfo) HasHeader() bool {
	return bi.Exists && bi.BlockStatus != StatusInvalid
}

// HasBody returns whether the block exists and has a valid body
func (bi *BlockInfo) HasBody() bool {
	return bi.Exists && bi.BlockStatus != StatusInvalid && bi.BlockStatus != StatusHeaderOnly
}

// Clone returns a clone of BlockInfo
func (bi *BlockInfo) Clone() *BlockInfo {
	return &BlockInfo{
		Exists:         bi.Exists,
		BlockStatus:    bi.BlockStatus.Clone(),
		BlueScore:      bi.BlueScore,
		BlueWork:       new(big.Int).Set(bi.BlueWork),
		SelectedParent: bi.SelectedParent,
		MergeSetBlues:  CloneHashes(bi.MergeSetBlues),
		MergeSetReds:   CloneHashes(bi.MergeSetReds),
	}
}
