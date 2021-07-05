package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"sort"
)

func (csm *consensusStateManager) ResolveVirtual(stagingArea *model.StagingArea) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "csm.ResolveVirtual")
	defer onEnd()

	tips, err := csm.consensusStateStore.Tips(stagingArea, csm.databaseContext)
	if err != nil {
		return err
	}

	var sortErr error
	sort.Slice(tips, func(i, j int) bool {
		selectedParent, err := csm.ghostdagManager.ChooseSelectedParent(stagingArea, tips[i], tips[j])
		if err != nil {
			sortErr = err
			return false
		}

		return selectedParent.Equal(tips[i])
	})
	if sortErr != nil {
		return sortErr
	}

	var selectedTip *externalapi.DomainHash
	for _, tip := range tips {
		blockStatus, _, err := csm.resolveBlockStatus(stagingArea, tip, true)
		if err != nil {
			return err
		}

		if blockStatus == externalapi.StatusUTXOValid {
			selectedTip = tip
			break
		}
	}

	if selectedTip == nil {
		log.Warnf("Non of the DAG tips are valid")
		return nil
	}

	_, _, err = csm.updateVirtual(stagingArea, selectedTip, tips)
	if err != nil {
		return err
	}

	return nil
}
