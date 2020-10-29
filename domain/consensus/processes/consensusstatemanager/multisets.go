package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"
)

func (csm *consensusStateManager) calculateMultiset(
	acceptanceData []*model.BlockAcceptanceData, blockGHOSTDAGData *model.BlockGHOSTDAGData) (model.Multiset, error) {

	selectedParentMultiset, err := csm.multisetStore.Get(csm.databaseContext, blockGHOSTDAGData.SelectedParent)
	if err != nil {
		return nil, err
	}

	multiset := selectedParentMultiset.Clone()

	for _, blockAcceptanceData := range acceptanceData {
		for _, transactionAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
			if !transactionAcceptanceData.IsAccepted {
				continue
			}

			transaction := transactionAcceptanceData.Transaction

			var err error
			err = addTransactionToMultiset(multiset, transaction, blockGHOSTDAGData.BlueScore)
			if err != nil {
				return nil, err
			}
		}
	}

	return multiset, nil
}

func addTransactionToMultiset(multiset model.Multiset, transaction *externalapi.DomainTransaction,
	blockBlueScore uint64) error {

	for _, input := range transaction.Inputs {
		err := removeUTXOFromMultiset(multiset, input.UTXOEntry, &input.PreviousOutpoint)
		if err != nil {
			return err
		}
	}

	for i, output := range transaction.Outputs {
		outpoint := &externalapi.DomainOutpoint{
			ID:    *hashserialization.TransactionID(transaction),
			Index: uint32(i),
		}
		utxoEntry := &externalapi.UTXOEntry{
			Amount:          output.Value,
			ScriptPublicKey: output.ScriptPublicKey,
			BlockBlueScore:  blockBlueScore,
			IsCoinbase:      false,
		}
		err := addUTXOToMultiset(multiset, utxoEntry, outpoint)
		if err != nil {
			return err
		}
	}

	return nil
}

func addUTXOToMultiset(multiset model.Multiset, entry *externalapi.UTXOEntry,
	outpoint *externalapi.DomainOutpoint) error {

	serializedUTXO, err := hashserialization.SerializeUTXO(entry, outpoint)
	if err != nil {
		return err
	}
	multiset.Add(serializedUTXO)

	return nil
}

func removeUTXOFromMultiset(multiset model.Multiset, entry *externalapi.UTXOEntry,
	outpoint *externalapi.DomainOutpoint) error {

	serializedUTXO, err := hashserialization.SerializeUTXO(entry, outpoint)
	if err != nil {
		return err
	}
	multiset.Remove(serializedUTXO)

	return nil
}
