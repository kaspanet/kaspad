package utxo

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo/utxoalgebra"
)

type readOnlyUTXOIteratorWithDiff struct {
	baseIterator model.ReadOnlyUTXOSetIterator
	diff         *model.UTXODiff

	currentOutpoint  *externalapi.DomainOutpoint
	currentUTXOEntry *externalapi.UTXOEntry
	currentErr       error

	toAddIterator model.ReadOnlyUTXOSetIterator
}

// IteratorWithDiff applies a UTXODiff to given utxo iterator
func IteratorWithDiff(iterator model.ReadOnlyUTXOSetIterator, diff *model.UTXODiff) (model.ReadOnlyUTXOSetIterator, error) {
	if iteratorWithDiff, ok := iterator.(*readOnlyUTXOIteratorWithDiff); ok {
		combinedDiff, err := utxoalgebra.WithDiff(iteratorWithDiff.diff, diff)
		if err != nil {
			return nil, err
		}

		return IteratorWithDiff(iteratorWithDiff.baseIterator, combinedDiff)
	}

	return &readOnlyUTXOIteratorWithDiff{
		baseIterator:  iterator,
		diff:          diff,
		toAddIterator: CollectionIterator(diff.ToAdd),
	}, nil
}

func (r *readOnlyUTXOIteratorWithDiff) Next() bool {
	for r.baseIterator.Next() { // keep looping until we reach an outpoint/entry pair that is not in r.diff.ToRemove
		r.currentOutpoint, r.currentUTXOEntry, r.currentErr = r.baseIterator.Get()
		if !utxoalgebra.CollectionContainsWithBlueScore(r.diff.ToRemove, r.currentOutpoint, r.currentUTXOEntry.BlockBlueScore) {
			return true
		}
	}

	if r.toAddIterator.Next() {
		r.currentOutpoint, r.currentUTXOEntry, r.currentErr = r.toAddIterator.Get()
		return true
	}

	return false
}

func (r *readOnlyUTXOIteratorWithDiff) Get() (outpoint *externalapi.DomainOutpoint, utxoEntry *externalapi.UTXOEntry, err error) {
	return r.currentOutpoint, r.currentUTXOEntry, r.currentErr
}
