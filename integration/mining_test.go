package integration

import (
	"math/rand"
	"testing"

	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

func solveBlock(t *testing.T, block *util.Block) *wire.MsgBlock {
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
