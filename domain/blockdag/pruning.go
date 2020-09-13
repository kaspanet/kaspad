package blockdag

// pruningDepth is used to determine the depth from the virtual block where the pruning point is set once updated.
// The pruningDepth is defined in a way that it's mathematically proven that a block
// in virtual.blockAtDepth(pruningDepth).anticone that is not in virtual.past will never be in virtual.past.
func (dag *BlockDAG) pruningDepth() uint64 {
	k := uint64(dag.Params.K)
	return 2*dag.FinalityInterval() + 4*mergeSetSizeLimit*k + 2*k + 2
}
