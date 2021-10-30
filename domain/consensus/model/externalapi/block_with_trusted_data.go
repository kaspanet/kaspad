package externalapi

// BlockWithTrustedData is a block with pre-filled data
// that is not validated by the consensus.
// This is used when bring the pruning point and its
// anticone on a pruned-headers node.
type BlockWithTrustedData struct {
	Block        *DomainBlock
	DAAScore     uint64
	DAAWindow    []*TrustedDataDataDAABlock
	GHOSTDAGData []*BlockGHOSTDAGDataHashPair
}

// TrustedDataDataDAABlock is a block that belongs to BlockWithTrustedData.DAAWindow
// TODO: Currently each trusted data block contains the entire set of blocks in its
// DAA window. There's a lot of duplications between DAA windows of trusted blocks.
// This duplication should be optimized out.
type TrustedDataDataDAABlock struct {
	Block        *DomainBlock
	GHOSTDAGData *BlockGHOSTDAGData
}

// BlockGHOSTDAGDataHashPair is a pair of a block hash and its ghostdag data
type BlockGHOSTDAGDataHashPair struct {
	Hash         *DomainHash
	GHOSTDAGData *BlockGHOSTDAGData
}
