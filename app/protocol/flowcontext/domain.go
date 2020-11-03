package flowcontext

import "github.com/kaspanet/kaspad/domain/blockdag"

// DAG returns the DAG associated to the flow context.
func (f *FlowContext) Domain() *blockdag.BlockDAG {
	return f.domain
}
