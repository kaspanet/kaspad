package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/pkg/errors"
)

func (bp *blockProcessor) validateAndInsertImportedPruningPoint(
	stagingArea *model.StagingArea, newPruningPointHash *externalapi.DomainHash) error {
	// 주어진 가지치기 지점이 예상되는 가지치기 지점인지 확인
	log.Info("Checking that the given pruning point is the expected pruning point")

	isValidPruningPoint, err := bp.pruningManager.IsValidPruningPoint(stagingArea, newPruningPointHash)
	if err != nil {
		return err
	}

	if !isValidPruningPoint {
		// 유효한 가지치기 지점이 아닙니다.
		return errors.Wrapf(ruleerrors.ErrUnexpectedPruningPoint, "%s is not a valid pruning point",
			newPruningPointHash)
	}

	arePruningPointsInValidChain, err := bp.pruningManager.ArePruningPointsInValidChain(stagingArea)
	if err != nil {
		return err
	}

	if !arePruningPointsInValidChain {
		// 가지치기 지점은 유효한 "+"제네시스 체인
		return errors.Wrapf(ruleerrors.ErrInvalidPruningPointsChain, "pruning points do not compose a valid "+
			"chain to genesis")
	}
	// 새로운 가지치기 지점에 따라 합의 상태 관리자 업데이트
	log.Infof("Updating consensus state manager according to the new pruning point %s", newPruningPointHash)
	err = bp.consensusStateManager.ImportPruningPointUTXOSet(stagingArea, newPruningPointHash)
	if err != nil {
		return err
	}

	err = bp.updateVirtualAcceptanceDataAfterImportingPruningPoint(stagingArea)
	if err != nil {
		return err
	}

	return nil
}
