package txindex

import (
	"sync"

	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
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

	log.Info("Reseting TX Index")

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

	//we iterate from pruningPoint up - this gurantees that newer accepting blocks overwrite older ones in the store mapping
	//we also do not collect data before pruning point, since relevent blockData is pruned (see `TO DO`` note at the top regarding archival nodes)
	selectedParentChainChanges, err := ti.domain.Consensus().GetVirtualSelectedParentChainFromBlock(pruningPoint)
	if err != nil {
		return err
	}

	ti.removeTXIDs(selectedParentChainChanges, len(selectedParentChainChanges.Removed))
	if err != nil {
		return err
	}

	ti.addTXIDs(selectedParentChainChanges, len(selectedParentChainChanges.Added))
	if err != nil {
		return err
	}

	err = ti.store.commitTxIDsWithoutTransaction()
	if err != nil {
		return err
	}

	ti.store.updateAndCommitPruningPointWithoutTransaction(pruningPoint)
	if err != nil {
		return err
	}

	ti.store.commitVirtualParentsWithoutTransaction(virtualInfo.ParentHashes)
	if err != nil {
		return err
	}

	ti.store.discardAllButPruningPoint()

	return nil
}

func (ti *TXIndex) isSynced() (bool, error) {

	txIndexVirtualParents, err := ti.store.getVirtualParents()
	if err != nil {
		if database.IsNotFoundError(err) {
			return false, nil
		}
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

	log.Info("Updating TX index with VirtualSelectedParentChainChanges: %+v", virtualChangeSet.VirtualSelectedParentChainChanges)

	err := ti.removeTXIDs(virtualChangeSet.VirtualSelectedParentChainChanges, 1000)
	if err != nil {
		return nil, err
	}

	err = ti.addTXIDs(virtualChangeSet.VirtualSelectedParentChainChanges, 1000)
	if err != nil {
		return nil, err
	}

	added, removed, _, _ := ti.store.stagedData()
	txIndexChanges := &TXAcceptanceChange{
		Added:   added,
		Removed: removed,
	}

	ti.store.updateVirtualParents(virtualChangeSet.VirtualParents)

	err = ti.store.commit()
	if err != nil {
		return nil, err
	}

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
					log.Tracef("TX index Adding: %d transactions", len(blockAcceptanceData.TransactionAcceptanceData))
					if transactionAcceptanceData.IsAccepted {
						ti.store.add(
							*consensushashing.TransactionID(transactionAcceptanceData.Transaction),
							addedChainBlock,
						)
					}
				}
			}
		}
		position += chunkSize
	}
	return nil
}

func (ti *TXIndex) removeTXIDs(selectedParentChainChanges *externalapi.SelectedChainPath, chunkSize int) error {
	position := 0
	for position < len(selectedParentChainChanges.Removed) {
		var chainBlocksChunk []*externalapi.DomainHash

		if position+chunkSize > len(selectedParentChainChanges.Removed) {
			chainBlocksChunk = selectedParentChainChanges.Removed[position:]
		} else {
			chainBlocksChunk = selectedParentChainChanges.Removed[position : position+chunkSize]
		}
		// We use chunks in order to avoid blocking consensus for too long
		// note: this might not be needed here, but unsure how kaspad handles pruning / when reset might be called.
		chainBlocksAcceptanceData, err := ti.domain.Consensus().GetBlocksAcceptanceData(chainBlocksChunk)
		if err != nil {
			return err
		}
		for i, removedChainBlock := range chainBlocksChunk {
			chainBlockAcceptanceData := chainBlocksAcceptanceData[i]
			for _, blockAcceptanceData := range chainBlockAcceptanceData {
				log.Tracef("TX index Removing: %d transactions", len(blockAcceptanceData.TransactionAcceptanceData))
				for _, transactionAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
					if transactionAcceptanceData.IsAccepted {
						ti.store.remove(
							*consensushashing.TransactionID(transactionAcceptanceData.Transaction),
							removedChainBlock,
						)
					}
				}
			}
		}
		position += chunkSize
	}
	return nil
}

// TXAcceptingBlockHash returns the accepting block hash for for the given txID
func (ti *TXIndex) TXAcceptingBlockHash(txID *externalapi.DomainTransactionID) (blockHash *externalapi.DomainHash, found bool, err error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "TXIndex.TXAcceptingBlockHash")
	defer onEnd()

	ti.mutex.Lock()
	defer ti.mutex.Unlock()

	acceptingBlockHash, found, err := ti.store.getTxAcceptingBlockHash(txID)
	if err != nil {
		return nil, false, err
	}
	return acceptingBlockHash, found, nil
}

//TO DO: Get Block from TxID

//TO DO: Get Including BlockHash from AcceptingBlock

//TO DO: Get Including Block from AcceptingBlock

//TO DO: Get Number of confirmations
