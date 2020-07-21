package protocol

import "github.com/kaspanet/kaspad/blockdag"

func (m *Manager) DAG() *blockdag.BlockDAG {
	return m.dag
}
