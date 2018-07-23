package blockdag

import (
	"testing"
	"time"

	"github.com/daglabs/btcd/dagconfig"
)

func TestBlues(t *testing.T) {
	netParams := &dagconfig.SimNetParams

	blockVersion := int32(0x20000000)

	data := []struct {
		parents []string
		id      string
	}{
		{
			parents: []string{"A"},
			id:      "B",
		},
		{
			parents: []string{"A"},
			id:      "C",
		},
	}

	// Generate enough synthetic blocks for the rest of the test
	blockDag := newFakeDAG(netParams)
	genesisNode := blockDag.dag.SelectedTip()
	blockTime := genesisNode.Header().Timestamp
	blockIDMap := make(map[string]*blockNode)
	blockIDMap["A"] = genesisNode

	for _, blockData := range data {
		blockTime = blockTime.Add(time.Second)
		parents := blockSet{}
		for _, parentID := range blockData.parents {
			parent := blockIDMap[parentID]
			parents.add(parent)
		}
		node := newFakeNode(parents, blockVersion, 0, blockTime)
	}

	numBlocksToGenerate := uint32(5)
	for i := uint32(0); i < numBlocksToGenerate; i++ {
		blockTime = blockTime.Add(time.Second)
		node = newFakeNode(setFromSlice(node), blockVersion, 0, blockTime)
		chain.index.AddNode(node)
		chain.dag.SetTip(node)
	}
}
