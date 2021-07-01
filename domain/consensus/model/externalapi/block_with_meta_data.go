package externalapi

type BlockWithMetaData struct {
	Block        *DomainBlock
	DAAScore     uint64
	DAAWindow    []*DAABlock
	GHOSTDAGData []*BlockGHOSTDAGDataHashPair
}

type DAABlock struct {
	Header       BlockHeader
	GHOSTDAGData *BlockGHOSTDAGData
}

type BlockGHOSTDAGDataHashPair struct {
	Hash         *DomainHash
	GHOSTDAGData *BlockGHOSTDAGData
}
