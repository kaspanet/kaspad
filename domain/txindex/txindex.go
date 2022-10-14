package txindex

import (
	"fmt"
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

	if !isSynced || true {

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

	log.Tracef("Reseting TX Index")

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

	log.Tracef("Updating TX index with VirtualSelectedParentChainChanges: %+v", virtualChangeSet.VirtualSelectedParentChainChanges)

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
		for i, acceptingBlock := range chainBlocksChunk {
			chainBlockAcceptanceData := chainBlocksAcceptanceData[i]
			for _, blockAcceptanceData := range chainBlockAcceptanceData {
				for j, transactionAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
					log.Warnf("TX index Adding: %d transactions", len(blockAcceptanceData.TransactionAcceptanceData))
					if transactionAcceptanceData.IsAccepted {
						if err != nil {
							return err
						}
						transactionID := consensushashing.TransactionID(transactionAcceptanceData.Transaction)
						ti.store.add(
							*transactionID,
							uint32(j),                     //index of including block where transaction is found
							blockAcceptanceData.BlockHash, //this is the including block
							acceptingBlock,                //this is the accepting block
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
		for i, acceptingBlockHash := range chainBlocksChunk {
			chainBlockAcceptanceData := chainBlocksAcceptanceData[i]
			for _, blockAcceptanceData := range chainBlockAcceptanceData {
				log.Tracef("TX index Removing: %d transactions", len(blockAcceptanceData.TransactionAcceptanceData))
				for j, transactionAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
					if transactionAcceptanceData.IsAccepted {
						ti.store.remove(
							*consensushashing.TransactionID(transactionAcceptanceData.Transaction),
							uint32(j),
							blockAcceptanceData.BlockHash,
							acceptingBlockHash,
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
func (ti *TXIndex) TXAcceptingBlockHash(txID *externalapi.DomainTransactionID) (
	acceptingBlockHash *externalapi.DomainHash, found bool, err error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "TXIndex.TXAcceptingBlockHash")
	defer onEnd()

	ti.mutex.Lock()
	defer ti.mutex.Unlock()

	txData, found, err := ti.store.getTxData(txID)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}

	return txData.AcceptingBlockHash, found, nil
}

// TXAcceptingBlockHashes returns the accepting block hashes for for the given txIDs
func (ti *TXIndex) TXAcceptingBlockHashes(txIDs []*externalapi.DomainTransactionID) (
	txIDsToAcceptingBlockHashes TxIDsToBlockHashes, missingTxIds []*externalapi.DomainTransactionID, err error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "TXIndex.TXAcceptingBlockHashes")
	defer onEnd()

	ti.mutex.Lock()
	defer ti.mutex.Unlock()

	txIDsToTxIndexData, missingTxIds, err := ti.store.getTxsData(txIDs)
	if err != nil {
		return nil, nil, err
	}

	txIDsToAcceptingBlockHashes = make(TxIDsToBlockHashes)
	for txID, txIndexData := range txIDsToTxIndexData {
		txIDsToAcceptingBlockHashes[txID] = txIndexData.AcceptingBlockHash
	}

	return txIDsToAcceptingBlockHashes, missingTxIds, nil
}

// TXAcceptingBlock returns the accepting block for for the given txID
func (ti *TXIndex) TXAcceptingBlock(txID *externalapi.DomainTransactionID) (
	block *externalapi.DomainBlock, found bool, err error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "TXIndex.TXAcceptingBlock")
	defer onEnd()

	ti.mutex.Lock()
	defer ti.mutex.Unlock()

	txIndexData, found, err := ti.store.getTxData(txID)
	if err != nil {
		return nil, false, err
	}

	acceptingBlock, err := ti.domain.Consensus().GetBlockEvenIfHeaderOnly(txIndexData.AcceptingBlockHash)

	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, false, fmt.Errorf("accepting block %s missing for txID %s ", txIndexData.AcceptingBlockHash.String(), txID.String())
		}
		return nil, false, err
	}
	return acceptingBlock, true, nil
}

// TXAcceptingBlocks returns the accepting blocks for for the given txIDs
func (ti *TXIndex) TXAcceptingBlocks(txIDs []*externalapi.DomainTransactionID) (
	txIDsToAcceptingBlocks TxIDsToBlocks, notFound []*externalapi.DomainTransactionID, err error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "TXIndex.TXAcceptingBlocks")
	defer onEnd()

	ti.mutex.Lock()
	defer ti.mutex.Unlock()

	txIDsToTxIndexData, notFound, err := ti.store.getTxsData(txIDs)
	if err != nil {
		return nil, nil, err
	}

	txIDsToAcceptingBlocks = make(TxIDsToBlocks)

	for txID, txIndexData := range txIDsToTxIndexData {
		txIDsToAcceptingBlocks[txID], err = ti.domain.Consensus().GetBlockEvenIfHeaderOnly(txIndexData.AcceptingBlockHash)
		if err != nil {
			if database.IsNotFoundError(err) {
				return nil, nil, fmt.Errorf("accepting block %s missing for txID %s ", txIndexData.IncludingBlockHash.String(), txID.String())
			}
			return nil, notFound, err
		}
	}

	return txIDsToAcceptingBlocks, notFound, nil
}

// GetTX returns the domain transaction for for the given txID
func (ti *TXIndex) GetTX(txID *externalapi.DomainTransactionID) (
	tx *externalapi.DomainTransaction, found bool, err error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "TXIndex.GetTX")
	defer onEnd()

	ti.mutex.Lock()
	defer ti.mutex.Unlock()

	txIndexData, found, err := ti.store.getTxData(txID)
	if err != nil {
		return nil, false, err
	}

	acceptingBlock, err := ti.domain.Consensus().GetBlock(txIndexData.AcceptingBlockHash)
	if err != nil {
		return nil, false, err
	}

	return acceptingBlock.Transactions[txIndexData.IncludingIndex], true, nil
}

// GetTXs returns the domain transaction for for the given txIDs
func (ti *TXIndex) GetTXs(txIDs []*externalapi.DomainTransactionID) (
	txs []*externalapi.DomainTransaction, notFound []*externalapi.DomainTransactionID, err error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "TXIndex.GetTXs")
	defer onEnd()

	ti.mutex.Lock()
	defer ti.mutex.Unlock()

	txIDsToTxIndexData, notFound, err := ti.store.getTxsData(txIDs)
	if err != nil {
		return nil, nil, err
	}

	txs = make([]*externalapi.DomainTransaction, len(txIDsToTxIndexData))
	i := 0

	for txID, txIndexData := range txIDsToTxIndexData {
		includingBlock, err := ti.domain.Consensus().GetBlockEvenIfHeaderOnly(txIndexData.IncludingBlockHash)

		if err != nil {
			if database.IsNotFoundError(err) {
				return nil, nil, fmt.Errorf("including block %s missing for txID %s ", txIndexData.IncludingBlockHash.String(), txID.String())
			}
			return nil, nil, err
		}

		txs[i] = includingBlock.Transactions[txIndexData.IncludingIndex]
		i++
	}

	return txs, notFound, nil
}

// GetTXConfirmations returns the tx confirmations for for the given txID
func (ti *TXIndex) GetTXConfirmations(txID *externalapi.DomainTransactionID) (
	confirmations int64, found bool, err error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "TXIndex.GetTXConfirmations")
	defer onEnd()

	ti.mutex.Lock()
	defer ti.mutex.Unlock()

	txdata, found, err := ti.store.getTxData(txID)
	if err != nil {
		return 0, false, err
	}

	acceptingBlockHeader, err := ti.domain.Consensus().GetBlockHeader(txdata.AcceptingBlockHash)
	if err != nil {
		return -1, false, err
	}

	virtualBlock, err := ti.domain.Consensus().GetVirtualInfo()
	if err != nil {
		return 0, false, err
	}

	return int64(virtualBlock.BlueScore - acceptingBlockHeader.BlueScore()), true, nil
}

// GetTXsConfirmations returns the tx confirmations for for the given txIDs
func (ti *TXIndex) GetTXsConfirmations(txIDs []*externalapi.DomainTransactionID) (
	txIDsToConfirmations TxIDsToConfirmations, notFound []*externalapi.DomainTransactionID, err error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "TXIndex.GetTXsConfirmations")
	defer onEnd()

	ti.mutex.Lock()
	defer ti.mutex.Unlock()

	virtualBlock, err := ti.domain.Consensus().GetVirtualInfo()
	if err != nil {
		return nil, nil, err
	}

	txIDsToTxIndexData, _, err := ti.store.getTxsData(txIDs)
	if err != nil {
		return nil, nil, err
	}

	txIDsToConfirmations = make(TxIDsToConfirmations)
	for txID, txIndexData := range txIDsToTxIndexData {
		acceptingBlockHeader, err := ti.domain.Consensus().GetBlockHeader(txIndexData.AcceptingBlockHash)
		if err != nil {
			if database.IsNotFoundError(err) {
				return nil, nil, fmt.Errorf("including block %s missing for txID %s ", txIndexData.IncludingBlockHash.String(), txID.String())
			}
			return nil, nil, err
		}
		txIDsToConfirmations[txID] = int64(virtualBlock.BlueScore - acceptingBlockHeader.BlueScore())
	}

	return txIDsToConfirmations, notFound, nil
}

// TXIncludingBlockHash returns the including block hash for the given txID
func (ti *TXIndex) TXIncludingBlockHash(txID *externalapi.DomainTransactionID) (includingBlockHash *externalapi.DomainHash, found bool, err error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "TXIndex.TXIncludingBlockHash")
	defer onEnd()

	ti.mutex.Lock()
	defer ti.mutex.Unlock()

	txIndexData, found, err := ti.store.getTxData(txID)
	if err != nil {
		return nil, false, err
	}

	return txIndexData.IncludingBlockHash, true, nil
}

// TXIncludingBlock returns the including block hashes for for the given txIDs
func (ti *TXIndex) TXIncludingBlock(txID *externalapi.DomainTransactionID) (
	block *externalapi.DomainBlock, found bool, err error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "TXIndex.TXIncludingBlock")
	defer onEnd()

	ti.mutex.Lock()
	defer ti.mutex.Unlock()

	txIndexData, found, err := ti.store.getTxData(txID)
	if err != nil {
		return nil, false, err
	}

	includingBlock, err := ti.domain.Consensus().GetBlockEvenIfHeaderOnly(txIndexData.IncludingBlockHash)

	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, false, fmt.Errorf("including block %s missing for txID %s ", txIndexData.IncludingBlockHash.String(), txID.String())
		}
		return nil, false, err
	}
	return includingBlock, true, nil
}

// TXIncludingBlockHashes returns the including block hashes for for the given txI
func (ti *TXIndex) TXIncludingBlockHashes(txIDs []*externalapi.DomainTransactionID) (
	txIDsToIncludinglockHashes TxIDsToBlockHashes, missingTxIds []*externalapi.DomainTransactionID, err error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "TXIndex.TXIncludingBlockHashes")
	defer onEnd()

	ti.mutex.Lock()
	defer ti.mutex.Unlock()

	txIDsToTxIndexData, notFound, err := ti.store.getTxsData(txIDs)
	if err != nil {
		return nil, nil, err
	}

	txIDsToIncludinglockHashes = make(TxIDsToBlockHashes)

	for txID, txIndexData := range txIDsToTxIndexData {
		txIDsToIncludinglockHashes[txID] = txIndexData.IncludingBlockHash
	}

	return txIDsToIncludinglockHashes, notFound, nil
}

// TXIncludingBlocks returns the including block hashes for for the given txIDs
func (ti *TXIndex) TXIncludingBlocks(txIDs []*externalapi.DomainTransactionID) (
	txIDsToIncludingBlocks TxIDsToBlocks, notFound []*externalapi.DomainTransactionID, err error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "TXIndex.TXIncludingBlocks")
	defer onEnd()

	ti.mutex.Lock()
	defer ti.mutex.Unlock()

	txIDsToTxIndexData, notFound, err := ti.store.getTxsData(txIDs)
	if err != nil {
		return nil, nil, err
	}

	txIDsToIncludingBlocks = make(TxIDsToBlocks)

	for txID, txIndexData := range txIDsToTxIndexData {
		txIDsToIncludingBlocks[txID], err = ti.domain.Consensus().GetBlockEvenIfHeaderOnly(txIndexData.IncludingBlockHash)
		if err != nil {
			if database.IsNotFoundError(err) {
				return nil, nil, fmt.Errorf("including block %s missing for txID %s ", txIndexData.IncludingBlockHash.String(), txID.String())
			}
			return nil, nil, err
		}
	}

	return txIDsToIncludingBlocks, notFound, nil
}

// GetTXsBlueScores returns the tx's accepting bluescore for for the given txID
// Note: this is a optimization function to store and dynamically calc. tx confirmations with access to to virtual bluescore
// such as in the case of rpc confirmation notification listeners
func (ti *TXIndex) GetTXsBlueScores(txIDs []*externalapi.DomainTransactionID) (
	txIDsToBlueScores TxIDsToBlueScores, notFound []*externalapi.DomainTransactionID, err error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "TXIndex.GetTXsBlueScores")
	defer onEnd()

	ti.mutex.Lock()
	defer ti.mutex.Unlock()

	txIDsToTxIndexData, notFound, err := ti.store.getTxsData(txIDs)
	if err != nil {
		return nil, nil, err
	}

	txIDsToBlueScores = make(TxIDsToBlueScores)
	for txID, txIndexData := range txIDsToTxIndexData {
		acceptingBlockHeader, err := ti.domain.Consensus().GetBlockHeader(txIndexData.AcceptingBlockHash)
		if err != nil {
			if database.IsNotFoundError(err) {
				return nil, nil, fmt.Errorf("Accepting block %s missing for txID %s ", txIndexData.AcceptingBlockHash.String(), txID.String())
			}
			return nil, nil, err
		}
		txIDsToBlueScores[txID] = acceptingBlockHeader.BlueScore()
	}

	return txIDsToBlueScores, notFound, nil
}
