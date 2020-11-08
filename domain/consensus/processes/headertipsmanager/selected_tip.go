package headertipsmanager

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

func (h headerTipsManager) SelectedTip() (*externalapi.DomainHash, error) {
	tips, err := h.headerTipsStore.Tips(h.databaseContext)
	if err != nil {
		return nil, err
	}

	selectedTip, err := h.ghostdagManager.ChooseSelectedParent(tips...)
	if err != nil {
		return nil, err
	}

	return selectedTip, nil
}
