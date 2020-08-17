package blockdag

// pruningDepth is used to determine the depth from the virtual block where the pruning point is set once updated.
func (dag *BlockDAG) pruningDepth() uint64 {
	k := uint64(dag.Params.K)
	return 2*dag.FinalityInterval() + 4*mergeSetSizeLimit*k + 2*k + 2
}
