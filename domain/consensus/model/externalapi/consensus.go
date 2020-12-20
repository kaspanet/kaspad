package externalapi

// Consensus maintains the current core state of the node
type Consensus interface {
	BuildBlock(coinbaseData *DomainCoinbaseData, transactions []*DomainTransaction) (*DomainBlock, error)
	ValidateAndInsertBlock(block *DomainBlock) (*BlockInsertionResult, error)
	ValidateTransactionAndPopulateWithConsensusData(transaction *DomainTransaction) error

	GetBlock(blockHash *DomainHash) (*DomainBlock, error)
	GetBlockHeader(blockHash *DomainHash) (*DomainBlockHeader, error)
	GetBlockInfo(blockHash *DomainHash, options *BlockInfoOptions) (*BlockInfo, error)
	GetBlockAcceptanceData(blockHash *DomainHash) (AcceptanceData, error)

	GetHashesBetween(lowHash, highHash *DomainHash) ([]*DomainHash, error)
	GetMissingBlockBodyHashes(highHash *DomainHash) ([]*DomainHash, error)
	GetPruningPointUTXOSet(expectedPruningPointHash *DomainHash) ([]byte, error)
	ValidateAndInsertPruningPoint(newPruningPoint *DomainBlock, serializedUTXOSet []byte) error
	GetVirtualSelectedParent() (*DomainBlock, error)
	CreateBlockLocator(lowHash, highHash *DomainHash, limit uint32) (BlockLocator, error)
	FindNextBlockLocatorBoundaries(blockLocator BlockLocator) (lowHash, highHash *DomainHash, err error)
	GetSyncInfo() (*SyncInfo, error)
	Tips() ([]*DomainHash, error)
	GetVirtualInfo() (*VirtualInfo, error)
	GetVirtualSelectedParentChainFromBlock(blockHash *DomainHash) (*SelectedParentChainChanges, error)
}
