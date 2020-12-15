package externalapi

// BlockInfo contains various information about a specific block
type BlockInfo struct {
	Exists      bool
	BlockStatus BlockStatus
	BlueScore   uint64

	AcceptanceData AcceptanceData
}

// BlockInfoOptions are a set of options passed to GetBlockInfo
type BlockInfoOptions struct {
	IncludeAcceptanceData bool
}
