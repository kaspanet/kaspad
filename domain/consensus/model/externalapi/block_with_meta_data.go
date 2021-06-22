package externalapi

type BlockWithMetaData struct {
	Block        *DomainBlock
	DAAScore     uint64
	DAAWindow    []*BlockGHOSTDAGDataHeaderPair
	GHOSTDAGData []*BlockGHOSTDAGDataHashPair
}

type BlockGHOSTDAGDataHashPair struct {
	Hash         *DomainHash
	GHOSTDAGData *BlockGHOSTDAGData
}

type BlockGHOSTDAGDataHeaderPair struct {
	Header       BlockHeader
	GHOSTDAGData *BlockGHOSTDAGData
}
