package model

// DAGTopologyManager exposes methods for querying relationships
// between blocks in the DAG
type DAGTopologyManager interface {
	Parents(blockHash *DomainHash) ([]*DomainHash, error)
	Children(blockHash *DomainHash) ([]*DomainHash, error)
	IsParentOf(blockHashA *DomainHash, blockHashB *DomainHash) (bool, error)
	IsChildOf(blockHashA *DomainHash, blockHashB *DomainHash) (bool, error)
	IsAncestorOf(blockHashA *DomainHash, blockHashB *DomainHash) (bool, error)
	IsDescendantOf(blockHashA *DomainHash, blockHashB *DomainHash) (bool, error)
	IsAncestorOfAny(blockHash *DomainHash, potentialDescendants []*DomainHash) (bool, error)
	IsInSelectedParentChainOf(blockHashA *DomainHash, blockHashB *DomainHash) (bool, error)

	Tips() ([]*DomainHash, error)
	SetTips(tipHashes []*DomainHash) error
}
