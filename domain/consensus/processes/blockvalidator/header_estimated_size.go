package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// headerEstimatedSerializedSize is the estimated size of a block header in some
// serialization. This has to be deterministic, but not necessarily accurate, since
// it's only used to check block size limit violation.
func (v *blockValidator) headerEstimatedSerializedSize(header externalapi.BlockHeader) uint64 {
	size := uint64(0)
	size += 2 // Version (uint16)

	size += 8 // number of block levels (uint64)
	for _, blockLevelParents := range header.Parents() {
		size += 8                                                           // number of parents in the block level (uint64)
		size += uint64(externalapi.DomainHashSize * len(blockLevelParents)) // parents
	}

	size += externalapi.DomainHashSize // HashMerkleRoot
	size += externalapi.DomainHashSize // AcceptedIDMerkleRoot
	size += externalapi.DomainHashSize // UTXOCommitment
	size += 8                          // TimeInMilliseconds (int64)
	size += 4                          // Bits (uint32)
	size += 8                          // Nonce (uint64)

	return size
}
