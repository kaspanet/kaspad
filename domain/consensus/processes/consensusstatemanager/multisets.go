package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/multiset"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxoserialization"
)

func (csm *consensusStateManager) calculateMultiset(
	acceptanceData externalapi.AcceptanceData, blockGHOSTDAGData model.BlockGHOSTDAGData) (model.Multiset, error) {

	log.Debugf("calculateMultiset start for block with selected parent %s", blockGHOSTDAGData.SelectedParent())
	defer log.Debugf("calculateMultiset end for block with selected parent %s", blockGHOSTDAGData.SelectedParent())

	if blockGHOSTDAGData.SelectedParent() == nil {
		log.Debugf("Selected parent is nil, which could only happen for the genesis. " +
			"The genesis, by definition, has an empty multiset")
		return multiset.New(), nil
	}

	ms, err := csm.multisetStore.Get(csm.databaseContext, blockGHOSTDAGData.SelectedParent())
	if err != nil {
		return nil, err
	}
	log.Debugf("The multiset for the selected parent %s is: %s", blockGHOSTDAGData.SelectedParent(), ms.Hash())

	for _, blockAcceptanceData := range acceptanceData {
		for i, transactionAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
			transaction := transactionAcceptanceData.Transaction
			transactionID := consensushashing.TransactionID(transaction)
			if !transactionAcceptanceData.IsAccepted {
				log.Tracef("Skipping transaction %s because it was not accepted", transactionID)
				continue
			}

			isCoinbase := i == 0
			log.Tracef("Is transaction %s a coinbase transaction: %t", transactionID, isCoinbase)

			var err error
			err = addTransactionToMultiset(ms, transaction, blockGHOSTDAGData.BlueScore(), isCoinbase)
			if err != nil {
				return nil, err
			}
			log.Tracef("Added transaction %s to the multiset", transactionID)
		}
	}

	return ms, nil
}

func addTransactionToMultiset(multiset model.Multiset, transaction *externalapi.DomainTransaction,
	blockBlueScore uint64, isCoinbase bool) error {

	transactionID := consensushashing.TransactionID(transaction)
	log.Tracef("addTransactionToMultiset start for transaction %s", transactionID)
	defer log.Tracef("addTransactionToMultiset end for transaction %s", transactionID)

	for _, input := range transaction.Inputs {
		log.Tracef("Removing input %s at index %d from the multiset",
			input.PreviousOutpoint.TransactionID, input.PreviousOutpoint.Index)
		err := removeUTXOFromMultiset(multiset, input.UTXOEntry, &input.PreviousOutpoint)
		if err != nil {
			return err
		}
	}

	for i, output := range transaction.Outputs {
		outpoint := &externalapi.DomainOutpoint{
			TransactionID: *transactionID,
			Index:         uint32(i),
		}
		utxoEntry := utxo.NewUTXOEntry(output.Value, output.ScriptPublicKey, isCoinbase, blockBlueScore)

		log.Tracef("Adding input %s at index %d from the multiset", transactionID, i)
		err := addUTXOToMultiset(multiset, utxoEntry, outpoint)
		if err != nil {
			return err
		}
	}

	return nil
}

func addUTXOToMultiset(multiset model.Multiset, entry externalapi.UTXOEntry,
	outpoint *externalapi.DomainOutpoint) error {

	serializedUTXO, err := utxo.SerializeUTXO(entry, outpoint)
	if err != nil {
		return err
	}
	multiset.Add(serializedUTXO)

	return nil
}

func removeUTXOFromMultiset(multiset model.Multiset, entry externalapi.UTXOEntry,
	outpoint *externalapi.DomainOutpoint) error {

	serializedUTXO, err := utxo.SerializeUTXO(entry, outpoint)
	if err != nil {
		return err
	}
	multiset.Remove(serializedUTXO)

	return nil
}

func calcMultisetFromProtoUTXOSet(protoUTXOSet *utxoserialization.ProtoUTXOSet) (model.Multiset, error) {
	ms := multiset.New()
	for _, utxo := range protoUTXOSet.Utxos {
		ms.Add(utxo.EntryOutpointPair)
	}
	return ms, nil
}
