package integration

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/pow"
	"github.com/kaspanet/kaspad/util/difficulty"
	"math/rand"
	"testing"

	"github.com/kaspanet/kaspad/app/appmessage"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func solveBlock(block *externalapi.DomainBlock) *externalapi.DomainBlock {
	targetDifficulty := difficulty.CompactToBig(block.Header.Bits())
	headerForMining := block.Header.ToMutable()
	initialNonce := rand.Uint64()
	for i := initialNonce; i != initialNonce-1; i++ {
		headerForMining.SetNonce(i)
		if pow.CheckProofOfWorkWithTarget(headerForMining, targetDifficulty) {
			block.Header = headerForMining.ToImmutable()
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
