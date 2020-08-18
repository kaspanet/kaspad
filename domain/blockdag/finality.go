package blockdag

// FinalityInterval is the interval that determines the finality window of the DAG.
func (dag *BlockDAG) FinalityInterval() uint64 {
	return uint64(dag.Params.FinalityDuration / dag.Params.TargetTimePerBlock)
}
