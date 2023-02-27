package externalapi

// Consensus maintains the current core state of the node
type Consensus interface {
	Init(skipAddingGenesis bool) error
	BuildBlock(coinbaseData *DomainCoinbaseData, transactions []*DomainTransaction) (*DomainBlock, error)
	BuildBlockTemplate(coinbaseData *DomainCoinbaseData, transactions []*DomainTransaction) (*DomainBlockTemplate, error)
	ValidateAndInsertBlock(block *DomainBlock, updateVirtual bool) error
	ValidateAndInsertBlockWithTrustedData(block *BlockWithTrustedData, validateUTXO bool) error
	ValidateTransactionAndPopulateWithConsensusData(transaction *DomainTransaction) error
	ImportPruningPoints(pruningPoints []BlockHeader) error
	BuildPruningPointProof() (*PruningPointProof, error)
	ValidatePruningPointProof(pruningPointProof *PruningPointProof) error
	ApplyPruningPointProof(pruningPointProof *PruningPointProof) error

	GetBlock(blockHash *DomainHash) (*DomainBlock, bool, error)
	GetBlockEvenIfHeaderOnly(blockHash *DomainHash) (*DomainBlock, error)
	GetBlockHeader(blockHash *DomainHash) (BlockHeader, error)
	GetBlockInfo(blockHash *DomainHash) (*BlockInfo, error)
	GetBlockRelations(blockHash *DomainHash) (parents []*DomainHash, children []*DomainHash, err error)
	GetBlockAcceptanceData(blockHash *DomainHash) (AcceptanceData, error)
	GetBlocksAcceptanceData(blockHashes []*DomainHash) ([]AcceptanceData, error)

	GetHashesBetween(lowHash, highHash *DomainHash, maxBlocks uint64) (hashes []*DomainHash, actualHighHash *DomainHash, err error)
	GetAnticone(blockHash, contextHash *DomainHash, maxBlocks uint64) (hashes []*DomainHash, err error)
	GetMissingBlockBodyHashes(highHash *DomainHash) ([]*DomainHash, error)
	GetPruningPointUTXOs(expectedPruningPointHash *DomainHash, fromOutpoint *DomainOutpoint, limit int) ([]*OutpointAndUTXOEntryPair, error)
	GetVirtualUTXOs(expectedVirtualParents []*DomainHash, fromOutpoint *DomainOutpoint, limit int) ([]*OutpointAndUTXOEntryPair, error)
	PruningPoint() (*DomainHash, error)
	PruningPointHeaders() ([]BlockHeader, error)
	PruningPointAndItsAnticone() ([]*DomainHash, error)
	ClearImportedPruningPointData() error
	AppendImportedPruningPointUTXOs(outpointAndUTXOEntryPairs []*OutpointAndUTXOEntryPair) error
	ValidateAndInsertImportedPruningPoint(newPruningPoint *DomainHash) error
	GetVirtualSelectedParent() (*DomainHash, error)
	CreateBlockLocatorFromPruningPoint(highHash *DomainHash, limit uint32) (BlockLocator, error)
	CreateHeadersSelectedChainBlockLocator(lowHash, highHash *DomainHash) (BlockLocator, error)
	CreateFullHeadersSelectedChainBlockLocator() (BlockLocator, error)
	GetSyncInfo() (*SyncInfo, error)
	Tips() ([]*DomainHash, error)
	GetVirtualInfo() (*VirtualInfo, error)
	GetVirtualDAAScore() (uint64, error)
	IsValidPruningPoint(blockHash *DomainHash) (bool, error)
	ArePruningPointsViolatingFinality(pruningPoints []BlockHeader) (bool, error)
	GetVirtualSelectedParentChainFromBlock(blockHash *DomainHash) (*SelectedChainPath, error)
	IsInSelectedParentChainOf(blockHashA *DomainHash, blockHashB *DomainHash) (bool, error)
	GetHeadersSelectedTip() (*DomainHash, error)
	Anticone(blockHash *DomainHash) ([]*DomainHash, error)
	EstimateNetworkHashesPerSecond(startHash *DomainHash, windowSize int) (uint64, error)
	PopulateMass(transaction *DomainTransaction)
	ResolveVirtual(progressReportCallback func(uint64, uint64)) error
	BlockDAAWindowHashes(blockHash *DomainHash) ([]*DomainHash, error)
	TrustedDataDataDAAHeader(trustedBlockHash, daaBlockHash *DomainHash, daaBlockWindowIndex uint64) (*TrustedDataDataDAAHeader, error)
	TrustedBlockAssociatedGHOSTDAGDataBlockHashes(blockHash *DomainHash) ([]*DomainHash, error)
	TrustedGHOSTDAGData(blockHash *DomainHash) (*BlockGHOSTDAGData, error)
	IsChainBlock(blockHash *DomainHash) (bool, error)
	VirtualMergeDepthRoot() (*DomainHash, error)
	IsNearlySynced() (bool, error)
}
