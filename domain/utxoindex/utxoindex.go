package utxoindex

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

type UTXOIndex struct {
	consensus externalapi.Consensus
	store     *utxoIndexStore
}

func New(consensus externalapi.Consensus, database database.Database) *UTXOIndex {
	store := newUTXOIndexStore(database)
	return &UTXOIndex{
		consensus: consensus,
		store:     store,
	}
}

func (ui *UTXOIndex) Update(chainChanges *externalapi.SelectedParentChainChanges) error {
	for _, removedBlockHash := range chainChanges.Removed {
		err := ui.removeBlock(removedBlockHash)
		if err != nil {
			return err
		}
	}
	for _, addedBlockHash := range chainChanges.Added {
		err := ui.addBlock(addedBlockHash)
		if err != nil {
			return err
		}
	}
	return ui.store.commit()
}

func (ui *UTXOIndex) addBlock(blockHash *externalapi.DomainHash) error {
	blockInfo, err := ui.consensus.GetBlockInfo(blockHash, &externalapi.BlockInfoOptions{IncludeAcceptanceData: true})
	if err != nil {
		return err
	}
	for _, blockAcceptanceData := range blockInfo.AcceptanceData {
		for _, transactionAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
			ui.addTransaction(transactionAcceptanceData.Transaction, blockInfo.BlueScore)
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
			ui.removeTransaction(transactionAcceptanceData.Transaction)
		}
	}
	return nil
}

func (ui *UTXOIndex) addTransaction(transaction *externalapi.DomainTransaction, blockBlueScore uint64) {
	isCoinbase := transactionhelper.IsCoinBase(transaction)
	for _, transactionInput := range transaction.Inputs {
		ui.store.remove(transactionInput.UTXOEntry.ScriptPublicKey(), &transactionInput.PreviousOutpoint)
	}
	for index, transactionOutput := range transaction.Outputs {
		outpoint := externalapi.NewDomainOutpoint(transaction.ID, uint32(index))
		utxoEntry := utxo.NewUTXOEntry(transactionOutput.Value, transactionOutput.ScriptPublicKey, isCoinbase, blockBlueScore)
		ui.store.add(transactionOutput.ScriptPublicKey, outpoint, &utxoEntry)
	}
}

func (ui *UTXOIndex) removeTransaction(transaction *externalapi.DomainTransaction) {
	for index, transactionOutput := range transaction.Outputs {
		outpoint := externalapi.NewDomainOutpoint(transaction.ID, uint32(index))
		ui.store.remove(transactionOutput.ScriptPublicKey, outpoint)
	}
	for _, transactionInput := range transaction.Inputs {
		ui.store.add(transactionInput.UTXOEntry.ScriptPublicKey(), &transactionInput.PreviousOutpoint, &transactionInput.UTXOEntry)
	}
}
