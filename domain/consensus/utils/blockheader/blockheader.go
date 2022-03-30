package blockheader

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/pow"
	"math/big"
)

type blockHeader struct {
	version              uint16
	parents              []externalapi.BlockLevelParents
	hashMerkleRoot       *externalapi.DomainHash
	acceptedIDMerkleRoot *externalapi.DomainHash
	utxoCommitment       *externalapi.DomainHash
	timeInMilliseconds   int64
	bits                 uint32
	nonce                uint64
	daaScore             uint64
	blueScore            uint64
	blueWork             *big.Int
	pruningPoint         *externalapi.DomainHash

	isBlockLevelCached bool
	blockLevel         int
}

func (bh *blockHeader) BlueScore() uint64 {
	return bh.blueScore
}

func (bh *blockHeader) PruningPoint() *externalapi.DomainHash {
	return bh.pruningPoint
}

func (bh *blockHeader) DAAScore() uint64 {
	return bh.daaScore
}

func (bh *blockHeader) BlueWork() *big.Int {
	return bh.blueWork
}

func (bh *blockHeader) ToImmutable() externalapi.BlockHeader {
	return bh.clone()
}

func (bh *blockHeader) SetNonce(nonce uint64) {
	bh.isBlockLevelCached = false
	bh.nonce = nonce
}

func (bh *blockHeader) SetTimeInMilliseconds(timeInMilliseconds int64) {
	bh.isBlockLevelCached = false
	bh.timeInMilliseconds = timeInMilliseconds
}

func (bh *blockHeader) SetHashMerkleRoot(hashMerkleRoot *externalapi.DomainHash) {
	bh.isBlockLevelCached = false
	bh.hashMerkleRoot = hashMerkleRoot
}

func (bh *blockHeader) Version() uint16 {
	return bh.version
}

func (bh *blockHeader) Parents() []externalapi.BlockLevelParents {
	return bh.parents
}

func (bh *blockHeader) DirectParents() externalapi.BlockLevelParents {
	if len(bh.parents) == 0 {
		return externalapi.BlockLevelParents{}
	}

	return bh.parents[0]
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

func (bh *blockHeader) Equal(other externalapi.BaseBlockHeader) bool {
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

	if !externalapi.ParentsEqual(bh.parents, other.Parents()) {
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

	if bh.daaScore != other.DAAScore() {
		return false
	}

	if bh.blueScore != other.BlueScore() {
		return false
	}

	if bh.blueWork.Cmp(other.BlueWork()) != 0 {
		return false
	}

	if !bh.pruningPoint.Equal(other.PruningPoint()) {
		return false
	}

	return true
}

func (bh *blockHeader) clone() *blockHeader {
	return &blockHeader{
		version:              bh.version,
		parents:              externalapi.CloneParents(bh.parents),
		hashMerkleRoot:       bh.hashMerkleRoot,
		acceptedIDMerkleRoot: bh.acceptedIDMerkleRoot,
		utxoCommitment:       bh.utxoCommitment,
		timeInMilliseconds:   bh.timeInMilliseconds,
		bits:                 bh.bits,
		nonce:                bh.nonce,
		daaScore:             bh.daaScore,
		blueScore:            bh.blueScore,
		blueWork:             bh.blueWork,
		pruningPoint:         bh.pruningPoint,
	}
}

func (bh *blockHeader) ToMutable() externalapi.MutableBlockHeader {
	return bh.clone()
}

func (bh *blockHeader) BlockLevel(maxBlockLevel int) int {
	if !bh.isBlockLevelCached {
		bh.blockLevel = pow.BlockLevel(bh, maxBlockLevel)
		bh.isBlockLevelCached = true
	}

	return bh.blockLevel
}

// NewImmutableBlockHeader returns a new immutable header
func NewImmutableBlockHeader(
	version uint16,
	parents []externalapi.BlockLevelParents,
	hashMerkleRoot *externalapi.DomainHash,
	acceptedIDMerkleRoot *externalapi.DomainHash,
	utxoCommitment *externalapi.DomainHash,
	timeInMilliseconds int64,
	bits uint32,
	nonce uint64,
	daaScore uint64,
	blueScore uint64,
	blueWork *big.Int,
	pruningPoint *externalapi.DomainHash,
) externalapi.BlockHeader {
	return &blockHeader{
		version:              version,
		parents:              parents,
		hashMerkleRoot:       hashMerkleRoot,
		acceptedIDMerkleRoot: acceptedIDMerkleRoot,
		utxoCommitment:       utxoCommitment,
		timeInMilliseconds:   timeInMilliseconds,
		bits:                 bits,
		nonce:                nonce,
		daaScore:             daaScore,
		blueScore:            blueScore,
		blueWork:             blueWork,
		pruningPoint:         pruningPoint,
	}
}
