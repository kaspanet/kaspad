package headertipsmanager

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

func (h headerTipsManager) SelectedTip() (*externalapi.DomainHash, error) {
	tips, err := h.headerTipsStore.Tips(h.databaseContext)
	if err != nil {
		return nil, err
	}

	return h.ghostdagManager.ChooseSelectedParent(tips...)
}
