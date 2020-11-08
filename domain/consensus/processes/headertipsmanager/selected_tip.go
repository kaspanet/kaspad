package headertipsmanager

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

func (h headerTipsManager) SelectedTip() (*externalapi.DomainHash, bool, error) {
	tips, err := h.headerTipsStore.Tips(h.databaseContext)
	if err != nil {
		return nil, false, err
	}

	if len(tips) == 0 {
		return nil, false, nil
	}

	selectedTip, err := h.ghostdagManager.ChooseSelectedParent(tips...)
	if err != nil {
		return nil, false, err
	}

	return selectedTip, true, nil
}
