package externalapi

// Consensus maintains the current core state of the node
type Consensus interface {
	Init(shouldNotAddGenesis bool) error
	BuildBlock(coinbaseData *DomainCoinbaseData, transactions []*DomainTransaction) (*DomainBlock, error)
	ValidateAndInsertBlock(block *DomainBlock, shouldValidateAgainstUTXO bool) (*BlockInsertionResult, error)
	ValidateAndInsertBlockWithMetaData(block *BlockWithMetaData, validateUTXO bool) (*BlockInsertionResult, error)
	ValidateTransactionAndPopulateWithConsensusData(transaction *DomainTransaction) error

	GetBlock(blockHash *DomainHash) (*DomainBlock, error)
	GetBlockEvenIfHeaderOnly(blockHash *DomainHash) (*DomainBlock, error)
	GetBlockHeader(blockHash *DomainHash) (BlockHeader, error)
	GetBlockInfo(blockHash *DomainHash) (*BlockInfo, error)
	GetBlockRelations(blockHash *DomainHash) (parents []*DomainHash, selectedParent *DomainHash, children []*DomainHash, err error)
	GetBlockAcceptanceData(blockHash *DomainHash) (AcceptanceData, error)

	GetHashesBetween(lowHash, highHash *DomainHash, maxBlocks uint64) (hashes []*DomainHash, actualHighHash *DomainHash, err error)
	GetPruningPointUTXOs(expectedPruningPointHash *DomainHash, fromOutpoint *DomainOutpoint, limit int) ([]*OutpointAndUTXOEntryPair, error)
	GetVirtualUTXOs(expectedVirtualParents []*DomainHash, fromOutpoint *DomainOutpoint, limit int) ([]*OutpointAndUTXOEntryPair, error)
	PruningPoint() (*DomainHash, error)
	PruningPointAndItsAnticoneWithMetaData() ([]*BlockWithMetaData, error)
	ClearImportedPruningPointData() error
	AppendImportedPruningPointUTXOs(outpointAndUTXOEntryPairs []*OutpointAndUTXOEntryPair) error
	ValidateAndInsertImportedPruningPoint(newPruningPoint *DomainHash) error
	GetVirtualSelectedParent() (*DomainHash, error)
	CreateBlockLocator(lowHash, highHash *DomainHash, limit uint32) (BlockLocator, error)
	CreateHeadersSelectedChainBlockLocator(lowHash, highHash *DomainHash) (BlockLocator, error)
	CreateFullHeadersSelectedChainBlockLocator() (BlockLocator, error)
	GetSyncInfo() (*SyncInfo, error)
	Tips() ([]*DomainHash, error)
	GetVirtualInfo() (*VirtualInfo, error)
	GetVirtualDAAScore() (uint64, error)
	IsValidPruningPoint(blockHash *DomainHash) (bool, error)
	GetVirtualSelectedParentChainFromBlock(blockHash *DomainHash) (*SelectedChainPath, error)
	IsInSelectedParentChainOf(blockHashA *DomainHash, blockHashB *DomainHash) (bool, error)
	GetHeadersSelectedTip() (*DomainHash, error)
	Anticone(blockHash *DomainHash) ([]*DomainHash, error)
	EstimateNetworkHashesPerSecond(startHash *DomainHash, windowSize int) (uint64, error)
}

type ConsensusWrapper interface {
	Consensus() Consensus
}
