package utxoindex

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
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
func New(consensus externalapi.Consensus, database database.Database) *UTXOIndex {
	store := newUTXOIndexStore(database)
	return &UTXOIndex{
		consensus: consensus,
		store:     store,
	}
}

func (ui *UTXOIndex) Reset() error {
	panic("unimplemented")
}

// Update updates the UTXO index with the given DAG selected parent chain changes
func (ui *UTXOIndex) Update(chainChanges *externalapi.SelectedChainPath) (*UTXOChanges, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "UTXOIndex.Update")
	defer onEnd()

	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	log.Tracef("Updating UTXO index with chainChanges: %+v", chainChanges)
	for _, removedBlockHash := range chainChanges.Removed {
		err := ui.removeBlock(removedBlockHash)
		if err != nil {
			return nil, err
		}
	}
	for _, addedBlockHash := range chainChanges.Added {
		err := ui.addBlock(addedBlockHash)
		if err != nil {
			return nil, err
		}
	}

	added, removed := ui.store.stagedData()
	utxoIndexChanges := &UTXOChanges{
		Added:   added,
		Removed: removed,
	}

	err := ui.store.commit()
	if err != nil {
		return nil, err
	}

	log.Tracef("UTXO index updated with the UTXOChanged: %+v", utxoIndexChanges)
	return utxoIndexChanges, nil
}

func (ui *UTXOIndex) addBlock(blockHash *externalapi.DomainHash) error {
	log.Tracef("Adding block %s to UTXO index", blockHash)
	acceptanceData, err := ui.consensus.GetBlockAcceptanceData(blockHash)
	if err != nil {
		return err
	}
	blockInfo, err := ui.consensus.GetBlockInfo(blockHash)
	if err != nil {
		return err
	}
	for _, blockAcceptanceData := range acceptanceData {
		for _, transactionAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
			if !transactionAcceptanceData.IsAccepted {
				continue
			}
			err := ui.addTransaction(transactionAcceptanceData.Transaction,
				transactionAcceptanceData.TransactionInputUTXOEntries, blockInfo.BlueScore)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ui *UTXOIndex) removeBlock(blockHash *externalapi.DomainHash) error {
	log.Tracef("Removing block %s from UTXO index", blockHash)
	acceptanceData, err := ui.consensus.GetBlockAcceptanceData(blockHash)
	if err != nil {
		return err
	}
	for _, blockAcceptanceData := range acceptanceData {
		for _, transactionAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
			if !transactionAcceptanceData.IsAccepted {
				continue
			}
			err := ui.removeTransaction(transactionAcceptanceData.Transaction,
				transactionAcceptanceData.TransactionInputUTXOEntries)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ui *UTXOIndex) addTransaction(transaction *externalapi.DomainTransaction,
	transactionInputUTXOEntries []externalapi.UTXOEntry, blockBlueScore uint64) error {

	transactionID := consensushashing.TransactionID(transaction)
	log.Tracef("Adding transaction %s to UTXO index", transactionID)

	isCoinbase := transactionhelper.IsCoinBase(transaction)
	for i, transactionInput := range transaction.Inputs {
		log.Tracef("Removing outpoint %s:%d from UTXO index",
			transactionInput.PreviousOutpoint.TransactionID, transactionInput.PreviousOutpoint.Index)
		inputUTXOEntry := transactionInputUTXOEntries[i]
		err := ui.store.remove(inputUTXOEntry.ScriptPublicKey(), &transactionInput.PreviousOutpoint)
		if err != nil {
			return err
		}
	}
	for index, transactionOutput := range transaction.Outputs {
		log.Tracef("Adding outpoint %s:%d to UTXO index", transactionID, index)
		outpoint := externalapi.NewDomainOutpoint(transactionID, uint32(index))
		utxoEntry := utxo.NewUTXOEntry(transactionOutput.Value, transactionOutput.ScriptPublicKey, isCoinbase, blockBlueScore)
		err := ui.store.add(transactionOutput.ScriptPublicKey, outpoint, utxoEntry)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ui *UTXOIndex) removeTransaction(transaction *externalapi.DomainTransaction,
	transactionInputUTXOEntries []externalapi.UTXOEntry) error {

	transactionID := consensushashing.TransactionID(transaction)
	log.Tracef("Removing transaction %s from UTXO index", transactionID)
	for index, transactionOutput := range transaction.Outputs {
		log.Tracef("Removing outpoint %s:%d from UTXO index", transactionID, index)
		outpoint := externalapi.NewDomainOutpoint(transactionID, uint32(index))
		err := ui.store.remove(transactionOutput.ScriptPublicKey, outpoint)
		if err != nil {
			return err
		}
	}
	for i, transactionInput := range transaction.Inputs {
		log.Tracef("Adding outpoint %s:%d to UTXO index",
			transactionInput.PreviousOutpoint.TransactionID, transactionInput.PreviousOutpoint.Index)
		inputUTXOEntry := transactionInputUTXOEntries[i]
		err := ui.store.add(inputUTXOEntry.ScriptPublicKey(), &transactionInput.PreviousOutpoint, transactionInput.UTXOEntry)
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
