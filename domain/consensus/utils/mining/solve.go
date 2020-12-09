package mining

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/pow"
	"math"
	"math/rand"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	utilsMath "github.com/kaspanet/kaspad/domain/consensus/utils/math"
	"github.com/pkg/errors"
)

// SolveBlock increments the given block's nonce until it matches the difficulty requirements in its bits field
func SolveBlock(block *externalapi.DomainBlock, rd *rand.Rand) {
	targetDifficulty := utilsMath.CompactToBig(block.Header.Bits)

	for i := rd.Uint64(); i < math.MaxUint64; i++ {
		block.Header.Nonce = i
		if pow.CheckProofOfWorkWithTarget(block.Header, targetDifficulty) {
			return
		}
	}

	panic(errors.New("went over all the nonce space and couldn't find a single one that gives a valid block"))
}
