package model

import "github.com/kaspanet/kaspad/util/daghash"

// ReachabilityTree maintains a structure that allows to answer
// reachability queries in sub-linear time
type ReachabilityTree interface {
	IsReachabilityTreeAncestorOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
	IsDAGAncestorOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
	ReachabilityChangeset(blockHash *daghash.Hash, blockGHOSTDAGData *BlockGHOSTDAGData) *ReachabilityChangeset
}
