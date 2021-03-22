package blockvalidator

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
)

// ValidateHeaderInContext validates block headers in the context of the current
// consensus state
func (v *blockValidator) ValidateHeaderInContext(blockHash *externalapi.DomainHash) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "ValidateHeaderInContext")
	defer onEnd()

	header, err := v.blockHeaderStore.BlockHeader(v.databaseContext, nil, blockHash)
	if err != nil {
		return err
	}

	hasValidatedHeader, err := v.hasValidatedHeader(blockHash)
	if err != nil {
		return err
	}

	if !hasValidatedHeader {
		var logErr error
		log.Debug(logger.NewLogClosure(func() string {
			var ghostdagData *model.BlockGHOSTDAGData
			ghostdagData, logErr = v.ghostdagDataStore.Get(v.databaseContext, nil, blockHash)
			if err != nil {
				return ""
			}

			return fmt.Sprintf("block %s blue score is %d", blockHash, ghostdagData.BlueScore())
		}))

		if logErr != nil {
			return logErr
		}
	}

	err = v.validateMedianTime(header)
	if err != nil {
		return err
	}

	err = v.checkMergeSizeLimit(blockHash)
	if err != nil {
		return err
	}

	// If needed - calculate reachability data right before calling CheckBoundedMergeDepth,
	// since it's used to find a block's finality point.
	// This might not be required if this block's header has previously been received during
	// headers-first synchronization.
	hasReachabilityData, err := v.reachabilityStore.HasReachabilityData(v.databaseContext, nil, blockHash)
	if err != nil {
		return err
	}
	if !hasReachabilityData {
		err = v.reachabilityManager.AddBlock(nil, blockHash)
		if err != nil {
			return err
		}
	}

	err = v.mergeDepthManager.CheckBoundedMergeDepth(nil, blockHash)
	if err != nil {
		return err
	}

	return nil
}

func (v *blockValidator) hasValidatedHeader(blockHash *externalapi.DomainHash) (bool, error) {
	exists, err := v.blockStatusStore.Exists(v.databaseContext, nil, blockHash)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, nil
	}

	status, err := v.blockStatusStore.Get(v.databaseContext, nil, blockHash)
	if err != nil {
		return false, err
	}

	return status == externalapi.StatusHeaderOnly, nil
}

// checkParentsIncest validates that no parent is an ancestor of another parent
func (v *blockValidator) checkParentsIncest(stagingArea *model.StagingArea, header externalapi.BlockHeader) error {
	for _, parentA := range header.ParentHashes() {
		for _, parentB := range header.ParentHashes() {
			if parentA.Equal(parentB) {
				continue
			}

			isAAncestorOfB, err := v.dagTopologyManager.IsAncestorOf(stagingArea, parentA, parentB)
			if err != nil {
				return err
			}

			if isAAncestorOfB {
				return errors.Wrapf(ruleerrors.ErrInvalidParentsRelation, "parent %s is an "+
					"ancestor of another parent %s",
					parentA,
					parentB,
				)
			}
		}
	}
	return nil
}

func (v *blockValidator) validateMedianTime(header externalapi.BlockHeader) error {
	if len(header.ParentHashes()) == 0 {
		return nil
	}

	// Ensure the timestamp for the block header is not before the
	// median time of the last several blocks (medianTimeBlocks).
	hash := consensushashing.HeaderHash(header)
	pastMedianTime, err := v.pastMedianTimeManager.PastMedianTime(hash)
	if err != nil {
		return err
	}

	if header.TimeInMilliseconds() <= pastMedianTime {
		return errors.Wrapf(ruleerrors.ErrTimeTooOld, "block timestamp of %d is not after expected %d",
			header.TimeInMilliseconds(), pastMedianTime)
	}

	return nil
}

func (v *blockValidator) checkMergeSizeLimit(hash *externalapi.DomainHash) error {
	ghostdagData, err := v.ghostdagDataStore.Get(v.databaseContext, nil, hash)
	if err != nil {
		return err
	}

	mergeSetSize := len(ghostdagData.MergeSet())

	if uint64(mergeSetSize) > v.mergeSetSizeLimit {
		return errors.Wrapf(ruleerrors.ErrViolatingMergeLimit,
			"The block merges %d blocks > %d merge set size limit", mergeSetSize, v.mergeSetSizeLimit)
	}

	return nil
}
