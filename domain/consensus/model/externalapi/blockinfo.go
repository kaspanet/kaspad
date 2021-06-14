package externalapi

import "math/big"

// BlockInfo contains various information about a specific block
type BlockInfo struct {
	Exists      bool
	BlockStatus BlockStatus
	BlueScore   uint64
	BlueWork    *big.Int
}

// Clone returns a clone of BlockInfo
func (bi *BlockInfo) Clone() *BlockInfo {
	return &BlockInfo{
		Exists:      bi.Exists,
		BlockStatus: bi.BlockStatus.Clone(),
		BlueScore:   bi.BlueScore,
	}
}
