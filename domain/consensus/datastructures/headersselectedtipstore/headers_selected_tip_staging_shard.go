package headersselectedtipstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type headersSelectedTipStagingShard struct {
	store          *headerSelectedTipStore
	newSelectedTip *externalapi.DomainHash
}

func (hsts *headerSelectedTipStore) stagingShard(stagingArea *model.StagingArea) *headersSelectedTipStagingShard {
	return stagingArea.GetOrCreateShard(model.StagingShardIDHeadersSelectedTip, func() model.StagingShard {
		return &headersSelectedTipStagingShard{
			store:          hsts,
			newSelectedTip: nil,
		}
	}).(*headersSelectedTipStagingShard)
}

func (hstss *headersSelectedTipStagingShard) Commit(dbTx model.DBTransaction) error {
	if hstss.newSelectedTip == nil {
		return nil
	}

	selectedTipBytes, err := hstss.store.serializeHeadersSelectedTip(hstss.newSelectedTip)
	if err != nil {
		return err
	}
	err = dbTx.Put(headerSelectedTipKey, selectedTipBytes)
	if err != nil {
		return err
	}
	hstss.store.cache = hstss.newSelectedTip

	return nil
}

func (hstss *headersSelectedTipStagingShard) isStaged() bool {
	return hstss.newSelectedTip != nil
}
