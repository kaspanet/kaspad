package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ReachabilityTree maintains a structure that allows to answer
// reachability queries in sub-linear time
type ReachabilityTree interface {
	AddBlock(blockHash *externalapi.DomainHash) error
	IsReachabilityTreeAncestorOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error)
	IsDAGAncestorOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error)
	UpdateReindexRoot(selectedTip *externalapi.DomainHash) error
}
