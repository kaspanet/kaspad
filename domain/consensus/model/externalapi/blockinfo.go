package externalapi

// BlockInfo contains various information about a specific block
type BlockInfo struct {
	Exists      bool
	BlockStatus BlockStatus

	IsBlockInHeaderPruningPointFuture bool
}

// Clone returns a clone of BlockInfo
func (bi *BlockInfo) Clone() *BlockInfo {
	return &BlockInfo{
		Exists:                            bi.Exists,
		BlockStatus:                       bi.BlockStatus.Clone(),
		IsBlockInHeaderPruningPointFuture: bi.IsBlockInHeaderPruningPointFuture,
	}
}
