package externalapi

// BlockInfo contains various information about a specific block
type BlockInfo struct {
	Exists      bool
	BlockStatus *BlockStatus

	IsHeaderInPruningPointFutureAndVirtualPast bool
}
