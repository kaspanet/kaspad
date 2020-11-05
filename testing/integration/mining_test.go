package integration

import (
	"math/rand"
	"testing"

	"github.com/kaspanet/kaspad/app/appmessage"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"

	"github.com/kaspanet/kaspad/util"
)

func solveBlock(block *externalapi.DomainBlock) *externalapi.DomainBlock {
	targetDifficulty := util.CompactToBig(block.Header.Bits)
	initialNonce := rand.Uint64()
	for i := initialNonce; i != initialNonce-1; i++ {
		block.Header.Nonce = i
		hash := hashserialization.BlockHash(block)
		if hashes.ToBig(hash).Cmp(targetDifficulty) <= 0 {
			return block
		}
	}

	panic("Failed to solve block! This should never happen")
}

func mineNextBlock(t *testing.T, harness *appHarness) *externalapi.DomainBlock {
	blockTemplate, err := harness.rpcClient.GetBlockTemplate(harness.miningAddress)
	if err != nil {
		t.Fatalf("Error getting block template: %+v", err)
	}

	block := appmessage.MsgBlockToDomainBlock(blockTemplate.MsgBlock)

	solveBlock(block)

	err = harness.rpcClient.SubmitBlock(block)
	if err != nil {
		t.Fatalf("Error submitting block: %s", err)
	}

	return block
}
