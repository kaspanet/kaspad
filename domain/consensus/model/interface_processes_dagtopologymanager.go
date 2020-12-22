package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// DAGTopologyManager exposes methods for querying relationships
// between blocks in the DAG
type DAGTopologyManager interface {
	Parents(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	Children(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	IsParentOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error)
	IsChildOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error)
	IsAncestorOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error)
	IsDescendantOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error)
	IsAncestorOfAny(blockHash *externalapi.DomainHash, potentialDescendants []*externalapi.DomainHash) (bool, error)
	IsInSelectedParentChainOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error)
	ChildInSelectedParentChainOf(context, highHash *externalapi.DomainHash) (*externalapi.DomainHash, error)

	SetParents(blockHash *externalapi.DomainHash, parentHashes []*externalapi.DomainHash) error
}
