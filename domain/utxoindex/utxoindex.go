package utxoindex

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"sync"
)

// UTXOIndex maintains an index between transaction scriptPublicKeys
// and UTXOs
type UTXOIndex struct {
	consensus externalapi.Consensus
	store     *utxoIndexStore

	mutex sync.Mutex
}

// New creates a new UTXO index
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

	log.Tracef("Updating UTXO index with VirtualUTXODiff: %+v", blockInsertionResult.VirtualUTXODiff)
	err := ui.removeUTXOs(blockInsertionResult.VirtualUTXODiff.ToRemove())
	if err != nil {
		return nil, err
	}

	err = ui.addUTXOs(blockInsertionResult.VirtualUTXODiff.ToAdd())
	if err != nil {
		return nil, err
	}

	ui.store.updateVirtualParents(blockInsertionResult.VirtualParents)

	added, removed, _ := ui.store.stagedData()
	utxoIndexChanges := &UTXOChanges{
		Added:   added,
		Removed: removed,
	}

	err = ui.store.commit()
	if err != nil {
		return nil, err
	}

	log.Tracef("UTXO index updated with the UTXOChanged: %+v", utxoIndexChanges)
	return utxoIndexChanges, nil
}

func (ui *UTXOIndex) addUTXOs(toAdd externalapi.UTXOCollection) error {
	iterator := toAdd.Iterator()
	for ok := iterator.First(); ok; ok = iterator.Next() {
		outpoint, entry, err := iterator.Get()
		if err != nil {
			return err
		}

		log.Tracef("Adding outpoint %s to UTXO index", outpoint)
		err = ui.store.add(entry.ScriptPublicKey(), outpoint, entry)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ui *UTXOIndex) removeUTXOs(toRemove externalapi.UTXOCollection) error {
	iterator := toRemove.Iterator()
	for ok := iterator.First(); ok; ok = iterator.Next() {
		outpoint, entry, err := iterator.Get()
		if err != nil {
			return err
		}

		log.Tracef("Removing outpoint %s from UTXO index", outpoint)
		err = ui.store.remove(entry.ScriptPublicKey(), outpoint)
		if err != nil {
			return err
		}
	}
	return nil
}

// UTXOs returns all the UTXOs for the given scriptPublicKey
func (ui *UTXOIndex) UTXOs(scriptPublicKey *externalapi.ScriptPublicKey) (UTXOOutpointEntryPairs, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "UTXOIndex.UTXOs")
	defer onEnd()

	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	return ui.store.getUTXOOutpointEntryPairs(scriptPublicKey)
}
