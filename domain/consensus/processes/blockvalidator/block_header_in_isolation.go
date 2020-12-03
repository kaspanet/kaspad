package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/pkg/errors"
)

// ValidateHeaderInIsolation validates block headers in isolation from the current
// consensus state
func (v *blockValidator) ValidateHeaderInIsolation(blockHash *externalapi.DomainHash) error {
	header, err := v.blockHeaderStore.BlockHeader(v.databaseContext, blockHash)
	if err != nil {
		return err
	}

	err = v.checkParentsLimit(header)
	if err != nil {
		return err
	}

	return nil
}

func (v *blockValidator) checkParentsLimit(header *externalapi.DomainBlockHeader) error {
	hash := consensushashing.HeaderHash(header)
	if len(header.ParentHashes) == 0 && *hash != *v.genesisHash {
		return errors.Wrapf(ruleerrors.ErrNoParents, "block has no parents")
	}

	if uint64(len(header.ParentHashes)) > uint64(v.maxBlockParents) {
		return errors.Wrapf(ruleerrors.ErrTooManyParents, "block header has %d parents, but the maximum allowed amount "+
			"is %d", len(header.ParentHashes), v.maxBlockParents)
	}
	return nil
}
