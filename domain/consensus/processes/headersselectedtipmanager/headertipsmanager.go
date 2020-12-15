package headersselectedtipmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type headerTipsManager struct {
	databaseContext model.DBReader

	dagTopologyManager      model.DAGTopologyManager
	ghostdagManager         model.GHOSTDAGManager
	headersSelectedTipStore model.HeaderSelectedTipStore
}

// New instantiates a new HeadersSelectedTipManager
func New(databaseContext model.DBReader,
	dagTopologyManager model.DAGTopologyManager,
	ghostdagManager model.GHOSTDAGManager,
	headersSelectedTipStore model.HeaderSelectedTipStore) model.HeadersSelectedTipManager {

	return &headerTipsManager{
		databaseContext:         databaseContext,
		dagTopologyManager:      dagTopologyManager,
		ghostdagManager:         ghostdagManager,
		headersSelectedTipStore: headersSelectedTipStore,
	}
}

func (h *headerTipsManager) AddHeaderTip(hash *externalapi.DomainHash) error {
	hasSelectedTip, err := h.headersSelectedTipStore.Has(h.databaseContext)
	if err != nil {
		return err
	}

	if !hasSelectedTip {
		h.headersSelectedTipStore.Stage(hash)
	} else {
		headersSelectedTip, err := h.headersSelectedTipStore.HeadersSelectedTip(h.databaseContext)
		if err != nil {
			return err
		}

		newHeadersSelectedTip, err := h.ghostdagManager.ChooseSelectedParent(headersSelectedTip, hash)
		if err != nil {
			return err
		}

		if *newHeadersSelectedTip != *headersSelectedTip {
			h.headersSelectedTipStore.Stage(newHeadersSelectedTip)
		}
	}

	return nil
}
