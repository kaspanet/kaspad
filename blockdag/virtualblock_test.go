// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"testing"
	"github.com/daglabs/btcd/dagconfig"
	"reflect"
	"github.com/daglabs/btcd/wire"
)

func TestVirtualBlock(t *testing.T) {
	hashCounter := uint32(0)
	params := dagconfig.MainNetParams
	buildNode := func(parents blockSet) *blockNode {
		header := wire.BlockHeader{Nonce: hashCounter}
		hashCounter++

		return newBlockNode(&header, parents, params.K)
	}

	// Create a DAG as follows:
	// 0 <- 1 <- 2
	//  \
	//   <- 3 <- 5
	//  \    X
	//   <- 4 <- 6
	node0 := buildNode(setFromSlice())
	node1 := buildNode(setFromSlice(node0))
	node2 := buildNode(setFromSlice(node1))
	node3 := buildNode(setFromSlice(node0))
	node4 := buildNode(setFromSlice(node0))
	node5 := buildNode(setFromSlice(node3, node4))
	node6 := buildNode(setFromSlice(node3, node4))

	tests := []struct {
		name                string
		tipsToSet           []*blockNode
		tipsToAdd           []*blockNode
		expectedTips        blockSet
		expectedSelectedTip *blockNode
	}{
		{
			name:                "empty virtual",
			tipsToSet:           []*blockNode{},
			tipsToAdd:           []*blockNode{},
			expectedTips:        newSet(),
			expectedSelectedTip: nil,
		},
		{
			name:                "virtual with genesis tip",
			tipsToSet:           []*blockNode{node0},
			tipsToAdd:           []*blockNode{},
			expectedTips:        setFromSlice(node0),
			expectedSelectedTip: node0,
		},
		{
			name:                "virtual with genesis tip, add child of genesis",
			tipsToSet:           []*blockNode{node0},
			tipsToAdd:           []*blockNode{node1},
			expectedTips:        setFromSlice(node1),
			expectedSelectedTip: node1,
		},
		{
			name:                "empty virtual, add a full DAG",
			tipsToSet:           []*blockNode{},
			tipsToAdd:           []*blockNode{node0, node1, node2, node3, node4, node5, node6},
			expectedTips:        setFromSlice(node2, node5, node6),
			expectedSelectedTip: node2,
		},
	}

	for _, test := range tests {
		virtual := newVirtualBlock(nil, params.K)
		virtual.SetTips(setFromSlice(test.tipsToSet...))
		for _, tipToAdd := range test.tipsToAdd {
			virtual.AddTip(tipToAdd)
		}

		resultTips := virtual.Tips()
		if !reflect.DeepEqual(resultTips, test.expectedTips) {
			t.Errorf("unexpected tips in test \"%s\". "+
				"Expected: %v, got: %v.", test.name, test.expectedTips, resultTips)
		}

		resultSelectedTip := virtual.SelectedTip()
		if !reflect.DeepEqual(resultSelectedTip, test.expectedSelectedTip) {
			t.Errorf("unexpected selected tip in test \"%s\". "+
				"Expected: %v, got: %v.", test.name, test.expectedSelectedTip, resultSelectedTip)
		}
	}
}
