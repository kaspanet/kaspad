package externalapi

// BlockWithMetaData is a block with pre-filled meta data
type BlockWithMetaData struct {
	Block        *DomainBlock
	DAAScore     uint64
	DAAWindow    []*BlockWithMetaDataDAABlock
	GHOSTDAGData []*BlockGHOSTDAGDataHashPair
}

// BlockWithMetaDataDAABlock is a block that belongs to BlockWithMetaData.DAAWindow
type BlockWithMetaDataDAABlock struct {
	Header       BlockHeader
	GHOSTDAGData *BlockGHOSTDAGData
}

// BlockGHOSTDAGDataHashPair is a pair of a block hash and its ghostdag data
type BlockGHOSTDAGDataHashPair struct {
	Hash         *DomainHash
	GHOSTDAGData *BlockGHOSTDAGData
}
