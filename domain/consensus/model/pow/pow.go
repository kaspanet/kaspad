package pow

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/domain/consensus/utils/serialization"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"math/big"
)

// CheckProofOfWorkWithTarget check's if the block has a valid PoW according to the provided target
// it does not check if the difficulty itself is valid or less than the maximum for the appropriate network
func CheckProofOfWorkWithTarget(header *externalapi.DomainBlockHeader, target *big.Int) bool {
	// The block pow must be less than the claimed target
	powNum := calcPowValue(header)

	// The block hash must be less or equal than the claimed target.
	return powNum.Cmp(target) <= 0
}

// CheckProofOfWorkByBits check's if the block has a valid PoW according to its Bits field
// it does not check if the difficulty itself is valid or less than the maximum for the appropriate network
func CheckProofOfWorkByBits(header *externalapi.DomainBlockHeader) bool {
	return CheckProofOfWorkWithTarget(header, util.CompactToBig(header.Bits))
}

func calcPowValue(header *externalapi.DomainBlockHeader) *big.Int {
	// Zero out the time and nonce.
	timestamp, nonce := header.TimeInMilliseconds, header.Nonce
	header.TimeInMilliseconds, header.Nonce = 0, 0

	prePowHash := consensushashing.HeaderHash(header)
	header.TimeInMilliseconds, header.Nonce = timestamp, nonce

	// PRE_POW_HASH || TIME || 32 zero byte padding || NONCE
	writer := hashes.NewHashWriter()
	_, err := writer.Write(prePowHash[:])
	if err != nil {
		panic(errors.Wrap(err, "this should never happen. SHA256's digest should never return an error"))
	}
	err = serialization.WriteElement(writer, timestamp)
	if err != nil {
		panic(errors.Wrap(err, "this should never happen. SHA256's digest should never return an error"))
	}
	zeroes := [32]byte{}
	_, err = writer.Write(zeroes[:])
	if err != nil {
		panic(errors.Wrap(err, "this should never happen. SHA256's digest should never return an error"))
	}
	err = serialization.WriteElement(writer, nonce)
	if err != nil {
		panic(errors.Wrap(err, "this should never happen. SHA256's digest should never return an error"))
	}
	return hashes.ToBig(writer.Finalize())
}
