package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ReachabilityManager maintains a structure that allows to answer
// reachability queries in sub-linear time
type ReachabilityManager interface {
	AddBlock(blockHash *externalapi.DomainHash) error
	IsReachabilityTreeAncestorOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error)
	IsDAGAncestorOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error)
	UpdateReindexRoot(selectedTip *externalapi.DomainHash) error
	FindAncestorOfThisAmongChildrenOfOther(this, other *externalapi.DomainHash) (*externalapi.DomainHash, error)
}
