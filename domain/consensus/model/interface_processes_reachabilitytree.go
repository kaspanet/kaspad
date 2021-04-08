package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ReachabilityManager maintains a structure that allows to answer
// reachability queries in sub-linear time
type ReachabilityManager interface {
	AddBlock(stagingArea *StagingArea, blockHash *externalapi.DomainHash) error
	IsReachabilityTreeAncestorOf(stagingArea *StagingArea, blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error)
	IsDAGAncestorOf(stagingArea *StagingArea, blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error)
	UpdateReindexRoot(stagingArea *StagingArea, selectedTip *externalapi.DomainHash) error
	FindNextAncestor(stagingArea *StagingArea, descendant, ancestor *externalapi.DomainHash) (*externalapi.DomainHash, error)
}
