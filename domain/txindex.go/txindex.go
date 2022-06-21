package txindex

import (
	"sync"

	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/logger"
)

//TO DO: For archival nodes pruningPoint references should be substituted with the virtualchainBlock with the lowest bluescore!

// TXIndex maintains an index between transaction IDs and accepting block hashes
type TXIndex struct {
	domain domain.Domain
	store  *txIndexStore

	mutex sync.Mutex
}

// New creates a new TX index.
//
// NOTE: While this is called no new blocks can be added to the consensus.
func New(domain domain.Domain, database database.Database) (*TXIndex, error) {
	txIndex := &TXIndex{
		domain: domain,
		store:  newTXIndexStore(database),
	}
	isSynced, err := txIndex.isSynced()
	if err != nil {
		return nil, err
	}

	if !isSynced {

		err := txIndex.Reset()
		if err != nil {
			return nil, err
		}
	}

	return txIndex, nil
}

// Reset deletes the whole Txindex and resyncs it from consensus.
func (ti *TXIndex) Reset() error {
	ti.mutex.Lock()
	defer ti.mutex.Unlock()

	err := ti.store.deleteAll()
	if err != nil {
		return err
	}

	virtualInfo, err := ti.domain.Consensus().GetVirtualInfo()
	if err != nil {
		return err
	}

	pruningPoint, err := ti.domain.Consensus().PruningPoint()
	if err != nil {
		return err
	}

	const chunkSize = 1000

	//we iterate from pruningPoint up - this gurantees that newer accepting blocks overwrite older ones in the store mapping
	//we also do not collect data before pruning point, since relevent blockData is pruned (see `TO DO`` note at the top regarding archival nodes)
	selectedParentChainChanges, err := ti.domain.Consensus().GetVirtualSelectedParentChainFromBlock(pruningPoint)
	if err != nil {
		return err
	}

	ti.addTXIDs(selectedParentChainChanges, 1000)

	err = ti.store.CommitWithoutTransaction()
	if err != nil {
		return err
	}

	err = ti.store.updateAndCommitPruningPointWithoutTransaction(pruningPoint)
	if err != nil {
		return err
	}

	return ti.store.updateAndCommitVirtualParentsWithoutTransaction(virtualInfo.ParentHashes)

}

func (ti *TXIndex) isSynced() (bool, error) {

	txIndexVirtualParents, err := ti.store.getVirtualParents()
	if err != nil {
		return false, err
	}

	txIndexPruningPoint, err := ti.store.getPruningPoint()
	if err != nil {
		if database.IsNotFoundError(err) {
			return false, nil
		}
		return false, err
	}

	virtualInfo, err := ti.domain.Consensus().GetVirtualInfo()
	if err != nil {
		return false, err
	}

	PruningPoint, err := ti.domain.Consensus().PruningPoint()
	if err != nil {
		return false, err
	}

	return externalapi.HashesEqual(virtualInfo.ParentHashes, txIndexVirtualParents) || txIndexPruningPoint.Equal(PruningPoint), nil
}

// Update updates the TX index with the given DAG selected parent chain changes
func (ti *TXIndex) Update(virtualChangeSet *externalapi.VirtualChangeSet) (*TXAcceptanceChange, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "TXIndex.Update")
	defer onEnd()

	ti.mutex.Lock()
	defer ti.mutex.Unlock()

	log.Tracef("Updating TX index with VirtualSelectedParentChainChanges: %+v", virtualChangeSet.VirtualSelectedParentChainChanges)

	err := ti.addTXIDs(virtualChangeSet.VirtualSelectedParentChainChanges, 1000)
	if err != nil {
		return nil, err
	}

	ti.store.updateVirtualParents(virtualChangeSet.VirtualParents)

	added, _, _ := ti.store.stagedData()
	txIndexChanges := &TXAcceptanceChange{
		Added: added,
	}

	removed, err := ti.store.commitAndReturnRemoved()
	if err != nil {
		return nil, err
	}

	txIndexChanges.Removed = removed

	log.Tracef("TX index updated with the TXAcceptanceChange: %+v", txIndexChanges)
	return txIndexChanges, nil
}

func (ti *TXIndex) addTXIDs(selectedParentChainChanges *externalapi.SelectedChainPath, chunkSize int) error {
	position := 0
	for position < len(selectedParentChainChanges.Added) {
		var chainBlocksChunk []*externalapi.DomainHash
		if position+chunkSize > len(selectedParentChainChanges.Added) {
			chainBlocksChunk = selectedParentChainChanges.Added[position:]
		} else {
			chainBlocksChunk = selectedParentChainChanges.Added[position : position+chunkSize]
		}

		if position+chunkSize > len(selectedParentChainChanges.Added) {
			chainBlocksChunk = selectedParentChainChanges.Added[position:]
		} else {
			chainBlocksChunk = selectedParentChainChanges.Added[position : position+chunkSize]
		}
		// We use chunks in order to avoid blocking consensus for too long
		// note: this might not be needed here, but unsure how kaspad handles pruning / when reset might be called.
		chainBlocksAcceptanceData, err := ti.domain.Consensus().GetBlocksAcceptanceData(chainBlocksChunk)
		if err != nil {
			return err
		}
		for i, addedChainBlock := range chainBlocksChunk {
			chainBlockAcceptanceData := chainBlocksAcceptanceData[i]
			for _, blockAcceptanceData := range chainBlockAcceptanceData {
				for _, transactionAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
					if transactionAcceptanceData.IsAccepted {
						ti.store.add(*transactionAcceptanceData.Transaction.ID, addedChainBlock)
					}
				}
			}
		}
		position += chunkSize
	}
	return nil
}

// TXAcceptingBlockHash returns all the UTXOs for the given scriptPublicKey
func (ui *TXIndex) TXAcceptingBlockHash(txID *externalapi.DomainTransactionID) (*externalapi.DomainHash, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "TXIndex.TXAcceptingBlockHash")
	defer onEnd()

	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	return ui.store.getUTXOOutpointEntryPairs(scriptPublicKey)
}

func (ti *TXIndex) TXAcceptingBlock(txID *externalapi.DomainTransactionID) (externalapi.DomainHash, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "TXIndex.TXAcceptingBlock")
	defer onEnd()

	ti.mutex.Lock()
	defer ti.mutex.Unlock()

	return ti.store.getUTXOOutpointEntryPairs(scriptPublicKey)
}

//TO DO: Get Block from TxID

//TO DO: Get Including BlockHash from AcceptingBlock

//TO DO: Get Including Block from AcceptingBlock

//TO DO: Get Confirmations