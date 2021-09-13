package pow

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/domain/consensus/utils/serialization"
	"github.com/kaspanet/kaspad/util/difficulty"

	"github.com/pkg/errors"
	"math/big"
)

// CheckProofOfWorkWithTarget check's if the block has a valid PoW according to the provided target
// it does not check if the difficulty itself is valid or less than the maximum for the appropriate network
func CheckProofOfWorkWithTarget(header externalapi.MutableBlockHeader, target *big.Int) bool {
	// The block pow must be less than the claimed target
	powNum := CalculateProofOfWorkValue(header)

	// The block hash must be less or equal than the claimed target.
	return powNum.Cmp(target) <= 0
}

// CheckProofOfWorkByBits check's if the block has a valid PoW according to its Bits field
// it does not check if the difficulty itself is valid or less than the maximum for the appropriate network
func CheckProofOfWorkByBits(header externalapi.MutableBlockHeader) bool {
	return CheckProofOfWorkWithTarget(header, difficulty.CompactToBig(header.Bits()))
}

// CalculateProofOfWorkValue hashes the given header and returns its big.Int value
func CalculateProofOfWorkValue(header externalapi.MutableBlockHeader) *big.Int {
	// Zero out the time and nonce.
	timestamp, nonce := header.TimeInMilliseconds(), header.Nonce()
	header.SetTimeInMilliseconds(0)
	header.SetNonce(0)

	prePowHash := consensushashing.HeaderHash(header)
	header.SetTimeInMilliseconds(timestamp)
	header.SetNonce(nonce)

	// PRE_POW_HASH || TIME || 32 zero byte padding || NONCE
	writer := hashes.NewPoWHashWriter()
	writer.InfallibleWrite(prePowHash.ByteSlice())
	err := serialization.WriteElement(writer, timestamp)
	if err != nil {
		panic(errors.Wrap(err, "this should never happen. Hash digest should never return an error"))
	}
	zeroes := [32]byte{}
	writer.InfallibleWrite(zeroes[:])
	err = serialization.WriteElement(writer, nonce)
	if err != nil {
		panic(errors.Wrap(err, "this should never happen. Hash digest should never return an error"))
	}
	return toBig(writer.Finalize())
}

// ToBig converts a externalapi.DomainHash into a big.Int treated as a little endian string.
func toBig(hash *externalapi.DomainHash) *big.Int {
	// We treat the Hash as little-endian for PoW purposes, but the big package wants the bytes in big-endian, so reverse them.
	buf := hash.ByteSlice()
	blen := len(buf)
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}

	return new(big.Int).SetBytes(buf)
}

func BlockLevel(header externalapi.BlockHeader) int {
	proofOfWorkValue := CalculateProofOfWorkValue(header.ToMutable())
	for blockLevel := 0; ; blockLevel++ {
		if blockLevel == constants.MaxBlockLevel || proofOfWorkValue.Bit(blockLevel+1) != 0 {
			return blockLevel
		}
	}
}
