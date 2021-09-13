package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type consensusStateStagingShard struct {
	store                  *consensusStateStore
	tipsStaging            []*externalapi.DomainHash
	virtualUTXODiffStaging externalapi.UTXODiff
}

func (bs *consensusStateStore) stagingShard(stagingArea *model.StagingArea) *consensusStateStagingShard {
	return stagingArea.GetOrCreateShard(bs.shardID, func() model.StagingShard {
		return &consensusStateStagingShard{
			store:                  bs,
			tipsStaging:            nil,
			virtualUTXODiffStaging: nil,
		}
	}).(*consensusStateStagingShard)
}

func (csss *consensusStateStagingShard) Commit(dbTx model.DBTransaction) error {
	err := csss.commitTips(dbTx)
	if err != nil {
		return err
	}

	err = csss.commitVirtualUTXODiff(dbTx)
	if err != nil {
		return err
	}

	return nil
}

func (csss *consensusStateStagingShard) isStaged() bool {
	return csss.tipsStaging != nil || csss.virtualUTXODiffStaging != nil
}
