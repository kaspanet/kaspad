package utxo

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/pkg/errors"
)

type mutableUTXODiff struct {
	toAdd    utxoCollection
	toRemove utxoCollection

	immutableReferences []*immutableUTXODiff
}

// NewMutableUTXODiff creates an empty mutable UTXO-Diff
func NewMutableUTXODiff() externalapi.MutableUTXODiff {
	return newMutableUTXODiff()
}

func newMutableUTXODiff() *mutableUTXODiff {
	return &mutableUTXODiff{
		toAdd:    utxoCollection{},
		toRemove: utxoCollection{},
	}
}

func (mud *mutableUTXODiff) ToImmutable() externalapi.UTXODiff {
	immutableReference := &immutableUTXODiff{
		mutableUTXODiff: mud,
		isInvalidated:   false,
	}

	mud.immutableReferences = append(mud.immutableReferences, immutableReference)

	return immutableReference
}

func (mud *mutableUTXODiff) invalidateImmutableReferences() {
	for _, immutableReference := range mud.immutableReferences {
		immutableReference.isInvalidated = true
	}

	mud.immutableReferences = nil
}

func (mud *mutableUTXODiff) WithDiff(other externalapi.UTXODiff) (externalapi.UTXODiff, error) {
	o, ok := other.(*immutableUTXODiff)
	if !ok {
		return nil, errors.New("other is not of type *immutableUTXODiff")
	}

	result, err := withDiff(mud, o.mutableUTXODiff)
	if err != nil {
		return nil, err
	}

	return result.ToImmutable(), nil
}

func (mud *mutableUTXODiff) WithDiffInPlace(other externalapi.UTXODiff) error {
	o, ok := other.(*immutableUTXODiff)
	if !ok {
		return errors.New("other is not of type *immutableUTXODiff")
	}

	mud.invalidateImmutableReferences()

	return withDiffInPlace(mud, o.mutableUTXODiff)
}

func (mud *mutableUTXODiff) DiffFrom(other externalapi.UTXODiff) (externalapi.UTXODiff, error) {
	o, ok := other.(*immutableUTXODiff)
	if !ok {
		return nil, errors.New("other is not of type *immutableUTXODiff")
	}

	result, err := diffFrom(mud, o.mutableUTXODiff)
	if err != nil {
		return nil, err
	}

	return result.ToImmutable(), nil
}

func (mud *mutableUTXODiff) ToAdd() externalapi.UTXOCollection {
	return mud.toAdd
}

func (mud *mutableUTXODiff) ToRemove() externalapi.UTXOCollection {
	return mud.toRemove
}
func (mud *mutableUTXODiff) AddTransaction(transaction *externalapi.DomainTransaction, blockDAAScore uint64) error {
	mud.invalidateImmutableReferences()

	for _, input := range transaction.Inputs {
		err := mud.removeEntry(&input.PreviousOutpoint, input.UTXOEntry)
		if err != nil {
			return err
		}
	}

	isCoinbase := transactionhelper.IsCoinBase(transaction)
	transactionID := *consensushashing.TransactionID(transaction)
	for i, output := range transaction.Outputs {
		outpoint := &externalapi.DomainOutpoint{
			TransactionID: transactionID,
			Index:         uint32(i),
		}
		entry := NewUTXOEntry(output.Value, output.ScriptPublicKey, isCoinbase, blockDAAScore)

		err := mud.addEntry(outpoint, entry)
		if err != nil {
			return err
		}
	}

	return nil
}

func (mud *mutableUTXODiff) addEntry(outpoint *externalapi.DomainOutpoint, entry externalapi.UTXOEntry) error {
	if mud.toRemove.containsWithDAAScore(outpoint, entry.BlockDAAScore()) {
		mud.toRemove.remove(outpoint)
	} else if mud.toAdd.Contains(outpoint) {
		return errors.Errorf("AddEntry: Cannot add outpoint %s twice", outpoint)
	} else {
		mud.toAdd.add(outpoint, entry)
	}
	return nil
}

func (mud *mutableUTXODiff) removeEntry(outpoint *externalapi.DomainOutpoint, entry externalapi.UTXOEntry) error {
	if mud.toAdd.containsWithDAAScore(outpoint, entry.BlockDAAScore()) {
		mud.toAdd.remove(outpoint)
	} else if mud.toRemove.Contains(outpoint) {
		return errors.Errorf("removeEntry: Cannot remove outpoint %s twice", outpoint)
	} else {
		mud.toRemove.add(outpoint, entry)
	}
	return nil
}

func (mud *mutableUTXODiff) clone() *mutableUTXODiff {
	if mud == nil {
		return nil
	}

	return &mutableUTXODiff{
		toAdd:    mud.toAdd.Clone(),
		toRemove: mud.toRemove.Clone(),
	}
}

func (mud *mutableUTXODiff) String() string {
	return fmt.Sprintf("toAdd: %s; toRemove: %s", mud.toAdd, mud.toRemove)
}
