package model

// DAGTopologyManager exposes methods for querying relationships
// between blocks in the DAG
type DAGTopologyManager interface {
	Parents(blockHash *DomainHash) []*DomainHash
	Children(blockHash *DomainHash) []*DomainHash
	IsParentOf(blockHashA *DomainHash, blockHashB *DomainHash) bool
	IsChildOf(blockHashA *DomainHash, blockHashB *DomainHash) bool
	IsAncestorOf(blockHashA *DomainHash, blockHashB *DomainHash) bool
	IsDescendantOf(blockHashA *DomainHash, blockHashB *DomainHash) bool
}
