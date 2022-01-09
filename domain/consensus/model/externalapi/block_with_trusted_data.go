package externalapi

// BlockWithTrustedData is a block with pre-filled data
// that is not validated by the consensus.
// This is used when bring the pruning point and its
// anticone on a pruned-headers node.
type BlockWithTrustedData struct {
	Block        *DomainBlock
	DAAWindow    []*TrustedDataDataDAAHeader
	GHOSTDAGData []*BlockGHOSTDAGDataHashPair
}

// TrustedDataDataDAAHeader is a block that belongs to BlockWithTrustedData.DAAWindow
type TrustedDataDataDAAHeader struct {
	Header       BlockHeader
	GHOSTDAGData *BlockGHOSTDAGData
}

// BlockGHOSTDAGDataHashPair is a pair of a block hash and its ghostdag data
type BlockGHOSTDAGDataHashPair struct {
	Hash         *DomainHash
	GHOSTDAGData *BlockGHOSTDAGData
}
