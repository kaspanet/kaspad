package utxoindex

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
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

// Update updates the UTXO index with the given DAG selected parent chain changes
func (ui *UTXOIndex) Update(chainChanges *externalapi.SelectedParentChainChanges) (*UTXOChanges, error) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

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
	return utxoIndexChanges, nil
}

func (ui *UTXOIndex) addBlock(blockHash *externalapi.DomainHash) error {
	blockInfo, err := ui.consensus.GetBlockInfo(blockHash, &externalapi.BlockInfoOptions{IncludeAcceptanceData: true})
	if err != nil {
		return err
	}
	for _, blockAcceptanceData := range blockInfo.AcceptanceData {
		for _, transactionAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
			err := ui.addTransaction(transactionAcceptanceData.Transaction, blockInfo.BlueScore)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ui *UTXOIndex) removeBlock(blockHash *externalapi.DomainHash) error {
	blockInfo, err := ui.consensus.GetBlockInfo(blockHash, &externalapi.BlockInfoOptions{IncludeAcceptanceData: true})
	if err != nil {
		return err
	}
	for _, blockAcceptanceData := range blockInfo.AcceptanceData {
		for _, transactionAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
			err := ui.removeTransaction(transactionAcceptanceData.Transaction)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ui *UTXOIndex) addTransaction(transaction *externalapi.DomainTransaction, blockBlueScore uint64) error {
	isCoinbase := transactionhelper.IsCoinBase(transaction)
	for _, transactionInput := range transaction.Inputs {
		err := ui.store.remove(transactionInput.UTXOEntry.ScriptPublicKey(), &transactionInput.PreviousOutpoint)
		if err != nil {
			return err
		}
	}

	transactionID := consensushashing.TransactionID(transaction)
	for index, transactionOutput := range transaction.Outputs {
		outpoint := externalapi.NewDomainOutpoint(transactionID, uint32(index))
		utxoEntry := utxo.NewUTXOEntry(transactionOutput.Value, transactionOutput.ScriptPublicKey, isCoinbase, blockBlueScore)
		err := ui.store.add(transactionOutput.ScriptPublicKey, outpoint, &utxoEntry)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ui *UTXOIndex) removeTransaction(transaction *externalapi.DomainTransaction) error {
	transactionID := consensushashing.TransactionID(transaction)
	for index, transactionOutput := range transaction.Outputs {
		outpoint := externalapi.NewDomainOutpoint(transactionID, uint32(index))
		err := ui.store.remove(transactionOutput.ScriptPublicKey, outpoint)
		if err != nil {
			return err
		}
	}
	for _, transactionInput := range transaction.Inputs {
		err := ui.store.add(transactionInput.UTXOEntry.ScriptPublicKey(), &transactionInput.PreviousOutpoint, &transactionInput.UTXOEntry)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ui *UTXOIndex) UTXOs(scriptPublicKey []byte) (UTXOOutpointEntryPairs, error) {
	scriptPublicKeyHexString := ConvertScriptPublicKeyToHexString(scriptPublicKey)
	return ui.store.getUTXOOutpointEntryPairs(scriptPublicKeyHexString)
}
