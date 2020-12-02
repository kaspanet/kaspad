package externalapi

// BlockInfo contains various information about a specific block
type BlockInfo struct {
	Exists      bool
	BlockStatus BlockStatus
	BlueScore   uint64

	IsBlockInHeaderPruningPointFuture bool
}
