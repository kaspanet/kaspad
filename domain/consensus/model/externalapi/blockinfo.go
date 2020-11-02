package externalapi

type BlockInfo struct {
	Exists      bool
	BlockStatus *BlockStatus

	IsHeaderInPruningPointFutureAndVirtualPast bool
}
