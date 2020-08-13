package integration

import (
	"math/rand"
	"testing"

	clientpkg "github.com/kaspanet/kaspad/network/rpc/client"

	"github.com/kaspanet/kaspad/network/domainmessage"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
)

func solveBlock(block *util.Block) *domainmessage.MsgBlock {
	msgBlock := block.MsgBlock()
	targetDifficulty := util.CompactToBig(msgBlock.Header.Bits)
	initialNonce := rand.Uint64()
	for i := initialNonce; i != initialNonce-1; i++ {
		msgBlock.Header.Nonce = i
		hash := msgBlock.BlockHash()
		if daghash.HashToBig(hash).Cmp(targetDifficulty) <= 0 {
			return msgBlock
		}
	}

	panic("Failed to solve block! This should never happen")
}

func mineNextBlock(t *testing.T, harness *appHarness) *util.Block {
	blockTemplate, err := harness.rpcClient.GetBlockTemplate(harness.miningAddress, "")
	if err != nil {
		t.Fatalf("Error getting block template: %+v", err)
	}

	block, err := clientpkg.ConvertGetBlockTemplateResultToBlock(blockTemplate)
	if err != nil {
		t.Fatalf("Error parsing blockTemplate: %s", err)
	}

	solveBlock(block)

	err = harness.rpcClient.SubmitBlock(block, nil)
	if err != nil {
		t.Fatalf("Error submitting block: %s", err)
	}

	return block
}
