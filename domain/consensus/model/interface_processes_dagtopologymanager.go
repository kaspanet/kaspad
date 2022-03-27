package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// DAGTopologyManager exposes methods for querying relationships
// between blocks in the DAG
type DAGTopologyManager interface {
	Parents(stagingArea *StagingArea, blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	Children(stagingArea *StagingArea, blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	IsParentOf(stagingArea *StagingArea, blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error)
	IsChildOf(stagingArea *StagingArea, blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error)
	IsAncestorOf(stagingArea *StagingArea, blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error)
	IsAncestorOfAny(stagingArea *StagingArea, blockHash *externalapi.DomainHash, potentialDescendants []*externalapi.DomainHash) (bool, error)
	IsAnyAncestorOf(stagingArea *StagingArea, potentialAncestors []*externalapi.DomainHash, blockHash *externalapi.DomainHash) (bool, error)
	IsInSelectedParentChainOf(stagingArea *StagingArea, blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error)
	ChildInSelectedParentChainOf(stagingArea *StagingArea, lowHash, highHash *externalapi.DomainHash) (*externalapi.DomainHash, error)

	SetParents(stagingArea *StagingArea, blockHash *externalapi.DomainHash, parentHashes []*externalapi.DomainHash) error
}
