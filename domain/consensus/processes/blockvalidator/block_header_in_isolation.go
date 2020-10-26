package blockvalidator

import (
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"sort"
)

// ValidateHeaderInIsolation validates block headers in isolation from the current
// consensus state
func (v *blockValidator) ValidateHeaderInIsolation(blockHash *externalapi.DomainHash) error {
	block, err := v.blockStore.Block(v.databaseContext, blockHash)
	if err != nil {
		return err
	}

	header := block.Header
	err = v.checkProofOfWork(header)
	if err != nil {
		return err
	}

	err = v.checkParentsLimit(header)
	if err != nil {
		return err
	}

	err = checkBlockParentsOrder(header)
	if err != nil {
		return err
	}

	return nil
}

func (v *blockValidator) checkParentsLimit(header *externalapi.DomainBlockHeader) error {
	hash := hashserialization.HeaderHash(header)
	if len(header.ParentHashes) == 0 && *hash != *v.genesisHash {
		return errors.Wrapf(ruleerrors.ErrNoParents, "block has no parents")
	}

	const maxParents = 10
	if len(header.ParentHashes) > maxParents {
		return errors.Wrapf(ruleerrors.ErrTooManyParents, "block header has %d parents, but the maximum allowed amount "+
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
func (v *blockValidator) checkProofOfWork(header *externalapi.DomainBlockHeader) error {
	// The target difficulty must be larger than zero.
	target := util.CompactToBig(header.Bits)
	if target.Sign() <= 0 {
		return errors.Wrapf(ruleerrors.ErrUnexpectedDifficulty, "block target difficulty of %064x is too low",
			target)
	}

	// The target difficulty must be less than the maximum allowed.
	if target.Cmp(v.powMax) > 0 {
		str := fmt.Sprintf("block target difficulty of %064x is "+
			"higher than max of %064x", target, v.powMax)
		return errors.Wrapf(ruleerrors.ErrUnexpectedDifficulty, str)
	}

	// The block hash must be less than the claimed target unless the flag
	// to avoid proof of work checks is set.
	if !v.skipPoW {
		// The block hash must be less than the claimed target.
		hash := hashserialization.HeaderHash(header)
		hashNum := hashes.ToBig(hash)
		if hashNum.Cmp(target) > 0 {
			return errors.Wrapf(ruleerrors.ErrUnexpectedDifficulty, "block hash of %064x is higher than "+
				"expected max of %064x", hashNum, target)
		}
	}

	return nil
}

//checkBlockParentsOrder ensures that the block's parents are ordered by hash
func checkBlockParentsOrder(header *externalapi.DomainBlockHeader) error {
	sortedHashes := make([]*externalapi.DomainHash, len(header.ParentHashes))
	for i, hash := range header.ParentHashes {
		sortedHashes[i] = hash
	}

	isSorted := sort.SliceIsSorted(sortedHashes, func(i, j int) bool {
		return hashes.Less(sortedHashes[i], sortedHashes[j])
	})

	if !isSorted {
		return errors.Wrapf(ruleerrors.ErrWrongParentsOrder, "block parents are not ordered by hash")
	}

	return nil
}
