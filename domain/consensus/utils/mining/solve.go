package mining

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	utilsMath "github.com/kaspanet/kaspad/domain/consensus/utils/math"
	"github.com/pkg/errors"
	"math"
	"math/rand"
)

func SolveBlock(block *externalapi.DomainBlock, rd *rand.Rand) {
	targetDifficulty := utilsMath.CompactToBig(block.Header.Bits)

	for i := rd.Uint64(); i < math.MaxUint64; i++ {
		block.Header.Nonce = i
		hash := consensusserialization.BlockHash(block)
		if hashes.ToBig(hash).Cmp(targetDifficulty) <= 0 {
			return
		}
	}

	panic(errors.New("went over all the nonce space and couldn't find a single one that gives a valid block"))
}
