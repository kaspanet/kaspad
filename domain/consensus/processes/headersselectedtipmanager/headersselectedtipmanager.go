package headersselectedtipmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type headerTipsManager struct {
	databaseContext model.DBReader

	dagTopologyManager        model.DAGTopologyManager
	dagTraversalManager       model.DAGTraversalManager
	ghostdagManager           model.GHOSTDAGManager
	headersSelectedTipStore   model.HeaderSelectedTipStore
	headersSelectedChainStore model.HeadersSelectedChainStore
}

// New instantiates a new HeadersSelectedTipManager
func New(databaseContext model.DBReader,
	dagTopologyManager model.DAGTopologyManager,
	dagTraversalManager model.DAGTraversalManager,
	ghostdagManager model.GHOSTDAGManager,
	headersSelectedTipStore model.HeaderSelectedTipStore,
	headersSelectedChainStore model.HeadersSelectedChainStore) model.HeadersSelectedTipManager {

	return &headerTipsManager{
		databaseContext:           databaseContext,
		dagTopologyManager:        dagTopologyManager,
		dagTraversalManager:       dagTraversalManager,
		ghostdagManager:           ghostdagManager,
		headersSelectedTipStore:   headersSelectedTipStore,
		headersSelectedChainStore: headersSelectedChainStore,
	}
}

func (h *headerTipsManager) AddHeaderTip(stagingArea *model.StagingArea, hash *externalapi.DomainHash) error {
	hasSelectedTip, err := h.headersSelectedTipStore.Has(h.databaseContext, stagingArea)
	if err != nil {
		return err
	}

	if !hasSelectedTip {
		h.headersSelectedTipStore.Stage(stagingArea, hash)

		err := h.headersSelectedChainStore.Stage(h.databaseContext, stagingArea, &externalapi.SelectedChainPath{
			Added:   []*externalapi.DomainHash{hash},
			Removed: nil,
		})
		if err != nil {
			return err
		}
	} else {
		headersSelectedTip, err := h.headersSelectedTipStore.HeadersSelectedTip(h.databaseContext, stagingArea)
		if err != nil {
			return err
		}

		newHeadersSelectedTip, err := h.ghostdagManager.ChooseSelectedParent(stagingArea, headersSelectedTip, hash)
		if err != nil {
			return err
		}

		if !newHeadersSelectedTip.Equal(headersSelectedTip) {
			h.headersSelectedTipStore.Stage(stagingArea, newHeadersSelectedTip)

			chainChanges, err := h.dagTraversalManager.CalculateChainPath(stagingArea, headersSelectedTip, newHeadersSelectedTip)
			if err != nil {
				return err
			}

			err = h.headersSelectedChainStore.Stage(h.databaseContext, stagingArea, chainChanges)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
