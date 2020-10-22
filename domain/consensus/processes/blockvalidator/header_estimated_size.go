package validator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// headerEstimatedSerializedSize is the estimated size of a block header in some
// serialization. This has to be deterministic, but not necessarily accurate, since
// it's only used to check block size limit violation.
func (v *validator) headerEstimatedSerializedSize(header *externalapi.DomainBlockHeader) uint64 {
	size := uint64(0)
	size += 4 // Version (int32)

	size += 8                                                 // number of parents (uint64)
	size += uint64(model.HashSize * len(header.ParentHashes)) // parents

	size += model.HashSize // HashMerkleRoot
	size += model.HashSize // AcceptedIDMerkleRoot
	size += model.HashSize // UTXOCommitment
	size += 8              // TimeInMilliseconds (int64)
	size += 4              // Bits (uint32)
	size += 8              // Nonce (uint64)

	return size
}
