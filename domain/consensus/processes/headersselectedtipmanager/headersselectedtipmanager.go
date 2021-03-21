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

func (h *headerTipsManager) AddHeaderTip(hash *externalapi.DomainHash) error {
	hasSelectedTip, err := h.headersSelectedTipStore.Has(h.databaseContext, nil)
	if err != nil {
		return err
	}

	if !hasSelectedTip {
		h.headersSelectedTipStore.Stage(nil, hash)

		err := h.headersSelectedChainStore.Stage(h.databaseContext, nil, &externalapi.SelectedChainPath{
			Added:   []*externalapi.DomainHash{hash},
			Removed: nil,
		})
		if err != nil {
			return err
		}
	} else {
		headersSelectedTip, err := h.headersSelectedTipStore.HeadersSelectedTip(h.databaseContext, nil)
		if err != nil {
			return err
		}

		newHeadersSelectedTip, err := h.ghostdagManager.ChooseSelectedParent(headersSelectedTip, hash)
		if err != nil {
			return err
		}

		if !newHeadersSelectedTip.Equal(headersSelectedTip) {
			h.headersSelectedTipStore.Stage(nil, newHeadersSelectedTip)

			chainChanges, err := h.dagTraversalManager.CalculateChainPath(headersSelectedTip, newHeadersSelectedTip)
			if err != nil {
				return err
			}

			err = h.headersSelectedChainStore.Stage(h.databaseContext, nil, chainChanges)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
