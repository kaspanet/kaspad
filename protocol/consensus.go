package protocol

import "github.com/kaspanet/kaspad/blockdag"

// DAG returns the DAG associated with the manager.
func (m *Manager) DAG() *blockdag.BlockDAG {
	return m.dag
}
