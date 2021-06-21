package externalapi

type BlockWithMetaData struct {
	Block        *DomainBlock
	DAAScore     uint64
	GHOSTDAGData *BlockGHOSTDAGData
}
