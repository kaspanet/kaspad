package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ReachabilityManager maintains a structure that allows to answer
// reachability queries in sub-linear time
type ReachabilityManager interface {
	AddBlock(blockHash *externalapi.DomainHash) error
	IsReachabilityTreeAncestorOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error)
	IsDAGAncestorOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error)
	UpdateReindexRoot(selectedTip *externalapi.DomainHash) error
}

type TestReachabilityManager interface {
	ReachabilityManager
	SetReachabilityReindexWindow(reindexWindow uint64)
	SetReachabilityReindexSlack(reindexSlack uint64)
	ReachabilityReindexSlack() uint64
}
