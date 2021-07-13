package utxoindex

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
	"sync"
)

// UTXOIndex maintains an index between transaction scriptPublicKeys
// and UTXOs
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
	err := ui.store.deleteAll()
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

		err = ui.store.addAndCommitOutpointsWithoutTransaction(virtualUTXOs)
		if err != nil {
			return err
		}

		if len(virtualUTXOs) < step {
			break
		}

		fromOutpoint = virtualUTXOs[len(virtualUTXOs)-1].Outpoint
	}

	// This has to be done last to mark that the reset went smoothly and no reset has to be called next time.
	return ui.store.updateAndCommitVirtualParentsWithoutTransaction(virtualInfo.ParentHashes)
}

func (ui *UTXOIndex) isSynced() (bool, error) {
	utxoIndexVirtualParents, err := ui.store.getVirtualParents()
	if err != nil {
		if database.IsNotFoundError(err) {
			return false, nil
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

	virtualUtxoDiff := blockInsertionResult.VirtualUTXODiff
	uis := ui.store

	log.Tracef("Updating UTXO index with VirtualUTXODiff: %+v", virtualUtxoDiff)

	for _, iteration := range []struct {
		opName                     string
		utxoCollection             externalapi.UTXOCollection
		setToInclude, setToExclude AddressesUTXOMap
	}{
		{"Removing", virtualUtxoDiff.ToRemove(), uis.toRemove, uis.toAdd},
		{"Adding", virtualUtxoDiff.ToAdd(), uis.toAdd, uis.toRemove},
	} {
		iterator := iteration.utxoCollection.Iterator()
		opName := iteration.opName
		setToInclude := iteration.setToInclude
		for ok := iterator.First(); ok; ok = iterator.Next() {
			outpoint, entry, err := iterator.Get()
			if err != nil {
				return nil, err
			}

			log.Tracef("UTXO index: %s outpoint: %s", opName, outpoint)

			key := ConvertScriptPublicKeyToString(entry.ScriptPublicKey())
			log.Tracef("scriptPublicKey %s: %s outpoint %s:%d",
				key, iteration.opName, outpoint.TransactionID, outpoint.Index)

			// If the outpoint exists in the opposite set simply remove it from there and continue
			if utxoMapToExclude, ok := iteration.setToExclude[key]; ok {
				if _, ok := utxoMapToExclude[*outpoint]; ok {
					log.Tracef("Outpoint exists in %s set, deleting it from there: %s:%d",
						opName, outpoint.TransactionID, outpoint.Index)
					delete(utxoMapToExclude, *outpoint)
					continue
				}
			}

			// Create a UTXOMap entry in toInclude set if it doesn't exist
			if _, ok := setToInclude[key]; !ok {
				log.Tracef("Creating key in %s set: %s", opName, key)
				iteration.setToInclude[key] = make(UTXOMap)
			}

			// Return an error if the outpoint already exists in toInclude set
			utxoMapToInclude := setToInclude[key]
			if _, ok := utxoMapToInclude[*outpoint]; ok {
				return nil, errors.Errorf("Cannot add outpoint because it’s being added already: %s", outpoint)
			}

			// If we add to toRemove set, we add nil instead of UTXO entry
			valueToAdd := entry
			if &setToInclude == &uis.toRemove {
				valueToAdd = nil
			}
			utxoMapToInclude[*outpoint] = valueToAdd

			log.Tracef("Done %s outpoint %s:%d on scriptPublicKey %s",
				opName, outpoint.TransactionID, outpoint.Index, key)

		}
		iterator.Close()
	}

	ui.store.updateVirtualParents(blockInsertionResult.VirtualParents)

	added, removed, _ := ui.store.stagedData()

	err := ui.store.commit()
	if err != nil {
		return nil, err
	}

	utxoIndexChanges := &UTXOChanges{
		Added:   added,
		Removed: removed,
	}

	log.Tracef("UTXO index updated with the UTXOChanged: %+v", utxoIndexChanges)
	return utxoIndexChanges, nil
}

// UTXOs returns all the UTXOs for the given scriptPublicKey
func (ui *UTXOIndex) UTXOs(scriptPublicKey *externalapi.ScriptPublicKey) (UTXOMap, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "UTXOIndex.UTXOs")
	defer onEnd()

	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	return ui.store.getUTXOOutpointEntryPairs(scriptPublicKey)
}
