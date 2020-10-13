package validator

import (
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"
	"github.com/kaspanet/kaspad/util"
	"sort"
)

// ValidateHeaderInIsolation validates block headers in isolation from the current
// consensus state
func (bv *Validator) ValidateHeaderInIsolation(header *model.DomainBlockHeader) error {
	// Ensure the proof of work bits in the block header is in min/max range
	// and the block hash is less than the target value described by the
	// bits.
	err := bv.checkProofOfWork(header)
	if err != nil {
		return err
	}

	err = bv.checkParentsLimit(header)
	if err != nil {
		return err
	}

	err = checkBlockParentsOrder(header)
	if err != nil {
		return err
	}

	return nil
}

func (bv *Validator) checkParentsLimit(header *model.DomainBlockHeader) error {
	hash := hashserialization.HeaderHash(header)
	if len(header.ParentHashes) == 0 && *hash != *bv.genesisHash {
		return ruleerrors.Errorf(ruleerrors.ErrNoParents, "block has no parents")
	}

	const maxParents = 10
	if len(header.ParentHashes) > maxParents {
		return ruleerrors.Errorf(ruleerrors.ErrTooManyParents, "block header has %d parents, but the maximum allowed amount "+
			"is %d", len(header.ParentHashes), maxParents)
	}
	return nil
}

// checkProofOfWork ensures the block header bits which indicate the target
// difficulty is in min/max range and that the block hash is less than the
// target difficulty as claimed.
//
// The flags modify the behavior of this function as follows:
//  - BFNoPoWCheck: The check to ensure the block hash is less than the target
//    difficulty is not performed.
func (bv *Validator) checkProofOfWork(header *model.DomainBlockHeader) error {
	// The target difficulty must be larger than zero.
	target := util.CompactToBig(header.Bits)
	if target.Sign() <= 0 {
		return ruleerrors.Errorf(ruleerrors.ErrUnexpectedDifficulty, "block target difficulty of %064x is too low",
			target)
	}

	// The target difficulty must be less than the maximum allowed.
	if target.Cmp(bv.powMax) > 0 {
		str := fmt.Sprintf("block target difficulty of %064x is "+
			"higher than max of %064x", target, bv.powMax)
		return ruleerrors.Errorf(ruleerrors.ErrUnexpectedDifficulty, str)
	}

	// The block hash must be less than the claimed target unless the flag
	// to avoid proof of work checks is set.
	if !bv.skipPoW {
		// The block hash must be less than the claimed target.
		hash := hashserialization.HeaderHash(header)
		hashNum := hashes.ToBig(hash)
		if hashNum.Cmp(target) > 0 {
			return ruleerrors.Errorf(ruleerrors.ErrUnexpectedDifficulty, "block hash of %064x is higher than "+
				"expected max of %064x", hashNum, target)
		}
	}

	return nil
}

//checkBlockParentsOrder ensures that the block's parents are ordered by hash
func checkBlockParentsOrder(header *model.DomainBlockHeader) error {
	sortedHashes := make([]*model.DomainHash, len(header.ParentHashes))
	for i, hash := range header.ParentHashes {
		sortedHashes[i] = hash
	}

	isSorted := sort.SliceIsSorted(sortedHashes, func(i, j int) bool {
		return hashes.Less(sortedHashes[i], sortedHashes[j])
	})

	if !isSorted {
		return ruleerrors.Errorf(ruleerrors.ErrWrongParentsOrder, "block parents are not ordered by hash")
	}

	return nil
}
