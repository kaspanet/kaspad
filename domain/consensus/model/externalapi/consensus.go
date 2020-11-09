package externalapi

// Consensus maintains the current core state of the node
type Consensus interface {
	BuildBlock(coinbaseData *DomainCoinbaseData, transactions []*DomainTransaction) (*DomainBlock, error)
	ValidateAndInsertBlock(block *DomainBlock) error
	ValidateTransactionAndPopulateWithConsensusData(transaction *DomainTransaction) error

	GetBlock(blockHash *DomainHash) (*DomainBlock, error)
	GetBlockHeader(blockHash *DomainHash) (*DomainBlockHeader, error)
	GetBlockInfo(blockHash *DomainHash) (*BlockInfo, error)

	GetHashesBetween(lowHash, highHash *DomainHash) ([]*DomainHash, error)
	GetMissingBlockBodyHashes(highHash *DomainHash) ([]*DomainHash, error)
	GetPruningPointUTXOSet() ([]byte, error)
	SetPruningPointUTXOSet(serializedUTXOSet []byte) error
	GetVirtualSelectedParent() (*DomainBlock, error)
	CreateBlockLocator(lowHash, highHash *DomainHash) (BlockLocator, error)
	FindNextBlockLocatorBoundaries(blockLocator BlockLocator) (lowHash, highHash *DomainHash, err error)
	GetSyncInfo() (*SyncInfo, error)
}
