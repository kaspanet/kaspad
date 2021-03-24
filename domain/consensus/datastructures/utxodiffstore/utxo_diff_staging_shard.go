package utxodiffstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type utxoDiffStagingShard struct {
	store              *utxoDiffStore
	utxoDiffToAdd      map[externalapi.DomainHash]externalapi.UTXODiff
	utxoDiffChildToAdd map[externalapi.DomainHash]*externalapi.DomainHash
	toDelete           map[externalapi.DomainHash]struct{}
}

func (uds *utxoDiffStore) stagingShard(stagingArea *model.StagingArea) *utxoDiffStagingShard {
	return stagingArea.GetOrCreateShard("UTXODiffStore", func() model.StagingShard {
		return &utxoDiffStagingShard{
			store:              uds,
			utxoDiffToAdd:      make(map[externalapi.DomainHash]externalapi.UTXODiff),
			utxoDiffChildToAdd: make(map[externalapi.DomainHash]*externalapi.DomainHash),
			toDelete:           make(map[externalapi.DomainHash]struct{}),
		}
	}).(*utxoDiffStagingShard)
}

func (udss utxoDiffStagingShard) Commit(dbTx model.DBTransaction) error {
	for hash, utxoDiff := range udss.utxoDiffToAdd {
		utxoDiffBytes, err := udss.store.serializeUTXODiff(utxoDiff)
		if err != nil {
			return err
		}
		err = dbTx.Put(udss.store.utxoDiffHashAsKey(&hash), utxoDiffBytes)
		if err != nil {
			return err
		}
		udss.store.utxoDiffCache.Add(&hash, utxoDiff)
	}

	for hash, utxoDiffChild := range udss.utxoDiffChildToAdd {
		if utxoDiffChild == nil {
			continue
		}

		utxoDiffChildBytes, err := udss.store.serializeUTXODiffChild(utxoDiffChild)
		if err != nil {
			return err
		}
		err = dbTx.Put(udss.store.utxoDiffChildHashAsKey(&hash), utxoDiffChildBytes)
		if err != nil {
			return err
		}
		udss.store.utxoDiffChildCache.Add(&hash, utxoDiffChild)
	}

	for hash := range udss.toDelete {
		err := dbTx.Delete(udss.store.utxoDiffHashAsKey(&hash))
		if err != nil {
			return err
		}
		udss.store.utxoDiffCache.Remove(&hash)

		err = dbTx.Delete(udss.store.utxoDiffChildHashAsKey(&hash))
		if err != nil {
			return err
		}
		udss.store.utxoDiffChildCache.Remove(&hash)
	}

	return nil
}

func (udss utxoDiffStagingShard) isStaged() bool {
	return len(udss.utxoDiffToAdd) != 0 || len(udss.utxoDiffChildToAdd) != 0 || len(udss.toDelete) != 0
}
