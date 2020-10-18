package model

// ReachabilityTree maintains a structure that allows to answer
// reachability queries in sub-linear time
type ReachabilityTree interface {
	IsReachabilityTreeAncestorOf(blockHashA *DomainHash, blockHashB *DomainHash) (bool, error)
	IsDAGAncestorOf(blockHashA *DomainHash, blockHashB *DomainHash) (bool, error)
	ReachabilityChangeset(blockHash *DomainHash, blockGHOSTDAGData *BlockGHOSTDAGData) (*ReachabilityChangeset, error)
}
