package blockheader

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type blockHeader struct {
	version              int32
	parentHashes         []*externalapi.DomainHash
	hashMerkleRoot       *externalapi.DomainHash
	acceptedIDMerkleRoot *externalapi.DomainHash
	utxoCommitment       *externalapi.DomainHash
	timeInMilliseconds   int64
	bits                 uint32
	nonce                uint64
}

func (bh *blockHeader) ToImmutable() externalapi.BlockHeader {
	return bh.clone()
}

func (bh *blockHeader) SetNonce(nonce uint64) {
	bh.nonce = nonce
}

func (bh *blockHeader) SetTimeInMilliseconds(timeInMilliseconds int64) {
	bh.timeInMilliseconds = timeInMilliseconds
}

func (bh *blockHeader) Version() int32 {
	return bh.version
}

func (bh *blockHeader) ParentHashes() []*externalapi.DomainHash {
	return bh.parentHashes
}

func (bh *blockHeader) HashMerkleRoot() *externalapi.DomainHash {
	return bh.hashMerkleRoot
}

func (bh *blockHeader) AcceptedIDMerkleRoot() *externalapi.DomainHash {
	return bh.acceptedIDMerkleRoot
}

func (bh *blockHeader) UTXOCommitment() *externalapi.DomainHash {
	return bh.utxoCommitment
}

func (bh *blockHeader) TimeInMilliseconds() int64 {
	return bh.timeInMilliseconds
}

func (bh *blockHeader) Bits() uint32 {
	return bh.bits
}

func (bh *blockHeader) Nonce() uint64 {
	return bh.nonce
}

func (bh *blockHeader) Equal(other externalapi.BlockHeader) bool {
	if bh == nil || other == nil {
		return bh == other
	}

	// If only the underlying value of other is nil it'll
	// make `other == nil` return false, so we check it
	// explicitly.
	downcastedOther := other.(*blockHeader)
	if bh == nil || downcastedOther == nil {
		return bh == downcastedOther
	}

	if bh.version != other.Version() {
		return false
	}

	if !externalapi.HashesEqual(bh.parentHashes, other.ParentHashes()) {
		return false
	}

	if !bh.hashMerkleRoot.Equal(other.HashMerkleRoot()) {
		return false
	}

	if !bh.acceptedIDMerkleRoot.Equal(other.AcceptedIDMerkleRoot()) {
		return false
	}

	if !bh.utxoCommitment.Equal(other.UTXOCommitment()) {
		return false
	}

	if bh.timeInMilliseconds != other.TimeInMilliseconds() {
		return false
	}

	if bh.bits != other.Bits() {
		return false
	}

	if bh.nonce != other.Nonce() {
		return false
	}

	return true
}

func (bh *blockHeader) clone() *blockHeader {
	return &blockHeader{
		version:              bh.version,
		parentHashes:         externalapi.CloneHashes(bh.parentHashes),
		hashMerkleRoot:       bh.hashMerkleRoot,
		acceptedIDMerkleRoot: bh.acceptedIDMerkleRoot,
		utxoCommitment:       bh.utxoCommitment,
		timeInMilliseconds:   bh.timeInMilliseconds,
		bits:                 bh.bits,
		nonce:                bh.nonce,
	}
}

func (bh *blockHeader) ToMutable() externalapi.MutableBlockHeader {
	return bh.clone()
}

// NewImmutableBlockHeader returns a new immutable header
func NewImmutableBlockHeader(
	version int32,
	parentHashes []*externalapi.DomainHash,
	hashMerkleRoot *externalapi.DomainHash,
	acceptedIDMerkleRoot *externalapi.DomainHash,
	utxoCommitment *externalapi.DomainHash,
	timeInMilliseconds int64,
	bits uint32,
	nonce uint64,
) externalapi.BlockHeader {
	return &blockHeader{
		version:              version,
		parentHashes:         parentHashes,
		hashMerkleRoot:       hashMerkleRoot,
		acceptedIDMerkleRoot: acceptedIDMerkleRoot,
		utxoCommitment:       utxoCommitment,
		timeInMilliseconds:   timeInMilliseconds,
		bits:                 bits,
		nonce:                nonce,
	}
}
