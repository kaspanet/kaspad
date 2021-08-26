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
func (v *blockValidator) ValidateHeaderInContext(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, isBlockWithTrustedData bool) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "ValidateHeaderInContext")
	defer onEnd()

	header, err := v.blockHeaderStore.BlockHeader(v.databaseContext, stagingArea, blockHash)
	if err != nil {
		return err
	}

	hasValidatedHeader, err := v.hasValidatedHeader(stagingArea, blockHash)
	if err != nil {
		return err
	}

	if !hasValidatedHeader {
		var logErr error
		log.Debug(logger.NewLogClosure(func() string {
			var ghostdagData *externalapi.BlockGHOSTDAGData
			ghostdagData, logErr = v.ghostdagDataStore.Get(v.databaseContext, stagingArea, blockHash, false)
			if err != nil {
				return ""
			}

			return fmt.Sprintf("block %s blue score is %d", blockHash, ghostdagData.BlueScore())
		}))

		if logErr != nil {
			return logErr
		}
	}

	err = v.validateMedianTime(stagingArea, header)
	if err != nil {
		return err
	}

	err = v.checkMergeSizeLimit(stagingArea, blockHash)
	if err != nil {
		return err
	}

	// If needed - calculate reachability data right before calling CheckBoundedMergeDepth,
	// since it's used to find a block's finality point.
	// This might not be required if this block's header has previously been received during
	// headers-first synchronization.
	hasReachabilityData, err := v.reachabilityStore.HasReachabilityData(v.databaseContext, stagingArea, blockHash)
	if err != nil {
		return err
	}
	if !hasReachabilityData {
		err = v.reachabilityManager.AddBlock(stagingArea, blockHash)
		if err != nil {
			return err
		}
	}

	if !isBlockWithTrustedData {
		err = v.checkIndirectParents(stagingArea, header)
		if err != nil {
			return err
		}
	}

	err = v.mergeDepthManager.CheckBoundedMergeDepth(stagingArea, blockHash, isBlockWithTrustedData)
	if err != nil {
		return err
	}

	err = v.checkDAAScore(stagingArea, blockHash, header)
	if err != nil {
		return err
	}
	err = v.checkBlueWork(stagingArea, blockHash, header, isBlockWithTrustedData)
	if err != nil {
		return err
	}
	err = v.checkFinalityPoint(stagingArea, blockHash, header, isBlockWithTrustedData)
	if err != nil {
		return err
	}

	return nil
}

func (v *blockValidator) hasValidatedHeader(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (bool, error) {
	exists, err := v.blockStatusStore.Exists(v.databaseContext, stagingArea, blockHash)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, nil
	}

	status, err := v.blockStatusStore.Get(v.databaseContext, stagingArea, blockHash)
	if err != nil {
		return false, err
	}

	return status == externalapi.StatusHeaderOnly, nil
}

// checkParentsIncest validates that no parent is an ancestor of another parent
func (v *blockValidator) checkParentsIncest(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) error {
	parents, err := v.dagTopologyManager.Parents(stagingArea, blockHash)
	if err != nil {
		return err
	}

	for _, parentA := range parents {
		for _, parentB := range parents {
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

func (v *blockValidator) validateMedianTime(stagingArea *model.StagingArea, header externalapi.BlockHeader) error {
	if len(header.DirectParents()) == 0 {
		return nil
	}

	// Ensure the timestamp for the block header is not before the
	// median time of the last several blocks (medianTimeBlocks).
	hash := consensushashing.HeaderHash(header)
	pastMedianTime, err := v.pastMedianTimeManager.PastMedianTime(stagingArea, hash)
	if err != nil {
		return err
	}

	if header.TimeInMilliseconds() <= pastMedianTime {
		return errors.Wrapf(ruleerrors.ErrTimeTooOld, "block timestamp of %d is not after expected %d",
			header.TimeInMilliseconds(), pastMedianTime)
	}

	return nil
}

func (v *blockValidator) checkMergeSizeLimit(stagingArea *model.StagingArea, hash *externalapi.DomainHash) error {
	ghostdagData, err := v.ghostdagDataStore.Get(v.databaseContext, stagingArea, hash, false)
	if err != nil {
		return err
	}

	mergeSetSize := len(ghostdagData.MergeSetBlues()) + len(ghostdagData.MergeSetReds())

	if uint64(mergeSetSize) > v.mergeSetSizeLimit {
		return errors.Wrapf(ruleerrors.ErrViolatingMergeLimit,
			"The block merges %d blocks > %d merge set size limit", mergeSetSize, v.mergeSetSizeLimit)
	}

	return nil
}

func (v *blockValidator) checkIndirectParents(stagingArea *model.StagingArea, header externalapi.BlockHeader) error {
	expectedParents, err := v.blockParentBuilder.BuildParents(stagingArea, header.DirectParents())
	if err != nil {
		return err
	}

	areParentsEqual := externalapi.ParentsEqual(header.Parents(), expectedParents)
	if !areParentsEqual {
		return errors.Wrapf(ruleerrors.ErrUnexpectedParents, "unexpected indirect block parents")
	}
	return nil
}

func (v *blockValidator) checkDAAScore(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash,
	header externalapi.BlockHeader) error {

	expectedDAAScore, err := v.daaBlocksStore.DAAScore(v.databaseContext, stagingArea, blockHash)
	if err != nil {
		return err
	}
	if header.DAAScore() != expectedDAAScore {
		return errors.Wrapf(ruleerrors.ErrUnexpectedDAAScore, "block DAA score of %d is not the expected value of %d", header.DAAScore(), expectedDAAScore)
	}
	return nil
}

func (v *blockValidator) checkBlueWork(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash,
	header externalapi.BlockHeader, isBlockWithTrustedData bool) error {

	ghostdagData, err := v.ghostdagDataStore.Get(v.databaseContext, stagingArea, blockHash, isBlockWithTrustedData)
	if err != nil {
		return err
	}
	expectedBlueWork := ghostdagData.BlueWork()
	if header.BlueWork().Cmp(expectedBlueWork) != 0 {
		return errors.Wrapf(ruleerrors.ErrUnexpectedBlueWork, "block blue work of %d is not the expected value of %d", header.BlueWork(), expectedBlueWork)
	}
	return nil
}

func (v *blockValidator) checkFinalityPoint(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash,
	header externalapi.BlockHeader, isBlockWithTrustedData bool) error {

	expectedFinalityPoint, err := v.finalityManager.FinalityPoint(stagingArea, blockHash, isBlockWithTrustedData)
	if err != nil {
		return err
	}
	if expectedFinalityPoint.Equal(model.VirtualGenesisBlockHash) {
		return nil
	}

	if !header.FinalityPoint().Equal(expectedFinalityPoint) {
		return errors.Wrapf(ruleerrors.ErrUnexpectedFinalityPoint, "block finality point of %s is not the expected hash of %s", header.FinalityPoint(), expectedFinalityPoint)
	}
	return nil
}
