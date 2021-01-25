package externalapi

// Consensus maintains the current core state of the node
type Consensus interface {
	BuildBlock(coinbaseData *DomainCoinbaseData, transactions []*DomainTransaction) (*DomainBlock, error)
	ValidateAndInsertBlock(block *DomainBlock) (*BlockInsertionResult, error)
	ValidateTransactionAndPopulateWithConsensusData(transaction *DomainTransaction) error

	GetBlock(blockHash *DomainHash) (*DomainBlock, error)
	GetBlockHeader(blockHash *DomainHash) (BlockHeader, error)
	GetBlockInfo(blockHash *DomainHash) (*BlockInfo, error)
	GetBlockAcceptanceData(blockHash *DomainHash) (AcceptanceData, error)

	GetHashesBetween(lowHash, highHash *DomainHash, maxBlueScoreDifference uint64) ([]*DomainHash, error)
	GetMissingBlockBodyHashes(highHash *DomainHash) ([]*DomainHash, error)
	GetPruningPointUTXOs(expectedPruningPointHash *DomainHash, fromOutpoint *DomainOutpoint, limit int) ([]*OutpointAndUTXOEntryPair, error)
	PruningPoint() (*DomainHash, error)
	ClearImportedPruningPointData() error
	AppendImportedPruningPointUTXOs(outpointAndUTXOEntryPairs []*OutpointAndUTXOEntryPair) error
	ValidateAndInsertImportedPruningPoint(newPruningPoint *DomainBlock) error
	GetVirtualSelectedParent() (*DomainHash, error)
	CreateBlockLocator(lowHash, highHash *DomainHash, limit uint32) (BlockLocator, error)
	CreateHeadersSelectedChainBlockLocator(lowHash, highHash *DomainHash) (BlockLocator, error)
	CreateFullHeadersSelectedChainBlockLocator() (BlockLocator, error)
	GetSyncInfo() (*SyncInfo, error)
	Tips() ([]*DomainHash, error)
	GetVirtualInfo() (*VirtualInfo, error)
	IsValidPruningPoint(blockHash *DomainHash) (bool, error)
	GetVirtualSelectedParentChainFromBlock(blockHash *DomainHash) (*SelectedChainPath, error)
	IsInSelectedParentChainOf(blockHashA *DomainHash, blockHashB *DomainHash) (bool, error)
	GetHeadersSelectedTip() (*DomainHash, error)
}
