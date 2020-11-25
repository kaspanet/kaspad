package utxoalgebra

import (
	"reflect"

	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"

	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// DiffAddTransaction modifies the provided utxoDiff with provided transaction.
func DiffAddTransaction(utxoDiff *model.UTXODiff, transaction *externalapi.DomainTransaction, blockBlueScore uint64) error {
	for _, input := range transaction.Inputs {
		err := diffRemoveEntry(utxoDiff, &input.PreviousOutpoint, input.UTXOEntry)
		if err != nil {
			return err
		}
	}

	isCoinbase := transactionhelper.IsCoinBase(transaction)
	transactionID := *consensusserialization.TransactionID(transaction)
	for i, output := range transaction.Outputs {
		outpoint := &externalapi.DomainOutpoint{
			TransactionID: transactionID,
			Index:         uint32(i),
		}
		entry := &externalapi.UTXOEntry{
			Amount:          output.Value,
			ScriptPublicKey: output.ScriptPublicKey,
			BlockBlueScore:  blockBlueScore,
			IsCoinbase:      isCoinbase,
		}

		err := diffAddEntry(utxoDiff, outpoint, entry)
		if err != nil {
			return err
		}
	}

	return nil
}

func diffAddEntry(diff *model.UTXODiff, outpoint *externalapi.DomainOutpoint, entry *externalapi.UTXOEntry) error {
	if CollectionContainsWithBlueScore(diff.ToRemove, outpoint, entry.BlockBlueScore) {
		collectionRemove(diff.ToRemove, outpoint)
	} else if _, exists := diff.ToAdd[*outpoint]; exists {
		return errors.Errorf("AddEntry: Cannot add outpoint %s twice", outpoint)
	} else {
		collectionAdd(diff.ToAdd, outpoint, entry)
	}
	return nil
}

func diffRemoveEntry(diff *model.UTXODiff, outpoint *externalapi.DomainOutpoint, entry *externalapi.UTXOEntry) error {
	if CollectionContainsWithBlueScore(diff.ToAdd, outpoint, entry.BlockBlueScore) {
		collectionRemove(diff.ToAdd, outpoint)
	} else if _, exists := diff.ToRemove[*outpoint]; exists {
		return errors.Errorf("removeEntry: Cannot remove outpoint %s twice", outpoint)
	} else {
		collectionAdd(diff.ToRemove, outpoint, entry)
	}
	return nil
}

func diffEqual(this *model.UTXODiff, other *model.UTXODiff) bool {
	if this == nil || other == nil {
		return this == other
	}

	return reflect.DeepEqual(this.ToAdd, other.ToAdd) &&
		reflect.DeepEqual(this.ToRemove, other.ToRemove)
}
