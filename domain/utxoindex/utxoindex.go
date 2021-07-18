package utxoindex

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"sync"
)

// UTXOIndex maintains an index between batch scriptPublicKeys and UTXOs
type UTXOIndex struct {
	consensus externalapi.Consensus
	store     *utxoIndexStore

	mutex sync.Mutex
}

// New creates a new UTXO index.
//
// NOTE: While this is called no new blocks can be added to the consensus.
func New(consensus externalapi.Consensus, database database.Database) (*UTXOIndex, error) {
	utxoIndex := &UTXOIndex{
		consensus: consensus,
		store:     newUTXOIndexStore(database),
	}

	isSynced, err := utxoIndex.isSynced()
	if err != nil {
		return nil, err
	}

	if !isSynced {
		err = utxoIndex.Reset()
		if err != nil {
			return nil, err
		}
	}

	return utxoIndex, nil
}

// Reset deletes the whole UTXO index and resyncs it from consensus.
func (ui *UTXOIndex) Reset() error {
	uis := ui.store

	err := uis.StartBatch()
	if err != nil {
		return err
	}
	defer uis.RollbackUnlessClosed()

	err = uis.clear()
	if err != nil {
		return err
	}

	virtualInfo, err := ui.consensus.GetVirtualInfo()
	if err != nil {
		return err
	}

	var fromOutpoint *externalapi.DomainOutpoint
	for {
		const step = 1000
		virtualUTXOs, err := ui.consensus.GetVirtualUTXOs(virtualInfo.ParentHashes, fromOutpoint, step)
		if err != nil {
			return err
		}
		for _, pair := range virtualUTXOs {
			err = ui.store.put(pair)
			if err != nil {
				return err
			}
		}
		if len(virtualUTXOs) < step {
			break
		}
		fromOutpoint = virtualUTXOs[len(virtualUTXOs)-1].Outpoint
	}

	// This has to be done last to mark that the reset went smoothly and no reset has to be called next time.
	err = ui.store.putVirtualParents(virtualInfo.ParentHashes)
	if err != nil {
		return err
	}

	return uis.Commit()
}

func (ui *UTXOIndex) isSynced() (bool, error) {
	utxoIndexVirtualParents, err := ui.store.getVirtualParents()
	if err != nil {
		if database.IsNotFoundError(err) {
			err = nil
		}
		return false, err
	}

	virtualInfo, err := ui.consensus.GetVirtualInfo()
	if err != nil {
		return false, err
	}

	return externalapi.HashesEqual(virtualInfo.ParentHashes, utxoIndexVirtualParents), nil
}

// Update updates the UTXO index with the given DAG selected parent chain changes
func (ui *UTXOIndex) Update(blockInsertionResult *externalapi.BlockInsertionResult) (*UTXOChanges, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "UTXOIndex.Update")
	defer onEnd()

	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	diff := blockInsertionResult.VirtualUTXODiff
	uis := ui.store

	log.Tracef("Updating UTXO index with VirtualUTXODiff: %+v", diff)

	err := uis.StartBatch()
	if err != nil {
		return nil, err
	}
	defer uis.RollbackUnlessClosed()

	// Structure for NotifyUTXOsChanged
	changes := &UTXOChanges{Added: AddressesUTXOMap{}, Removed: AddressesUTXOMap{}}

	for _, iteration := range []struct {
		utxos      externalapi.UTXOCollection
		changesMap AddressesUTXOMap
		isToRemove bool
	}{
		{diff.ToRemove(), changes.Removed, true},
		{diff.ToAdd(), changes.Added, false},
	} {
		iterator := iteration.utxos.Iterator()
		changesMap := iteration.changesMap
		for ok := iterator.First(); ok; ok = iterator.Next() {
			outpoint, entry, err := iterator.Get()
			if err != nil {
				return nil, err
			}
			pair := &externalapi.OutpointAndUTXOEntryPair{
				Outpoint:  outpoint,
				UTXOEntry: entry,
			}
			if iteration.isToRemove {
				err = uis.delete(pair)
			} else {
				err = uis.put(pair)
			}
			if err != nil {
				return nil, err
			}

			// Filling changes structure
			scriptPublicKeyString := ConvertScriptPublicKeyToString(entry.ScriptPublicKey())
			if changesMap[scriptPublicKeyString] == nil {
				changesMap[scriptPublicKeyString] = make(UTXOMap)
			}
			changesMap[scriptPublicKeyString][*outpoint] = entry
		}
	}

	err = uis.putVirtualParents(blockInsertionResult.VirtualParents)
	if err != nil {
		return nil, err
	}

	err = uis.Commit()
	if err != nil {
		return nil, err
	}

	return changes, nil
}

// GetUTXOsByScriptPublicKey returns all the UTXOs for the given scriptPublicKey
func (ui *UTXOIndex) GetUTXOsByScriptPublicKey(scriptPublicKey *externalapi.ScriptPublicKey) (UTXOMap, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "UTXOIndex.UTXOs")
	defer onEnd()

	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	return ui.store.getUTXOsByScriptPublicKey(scriptPublicKey)
}
