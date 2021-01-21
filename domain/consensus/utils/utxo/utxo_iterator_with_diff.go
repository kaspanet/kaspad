package utxo

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

type readOnlyUTXOIteratorWithDiff struct {
	baseIterator model.ReadOnlyUTXOSetIterator
	diff         *immutableUTXODiff

	currentOutpoint  *externalapi.DomainOutpoint
	currentUTXOEntry externalapi.UTXOEntry
	currentErr       error

	toAddIterator model.ReadOnlyUTXOSetIterator
}

// IteratorWithDiff applies a UTXODiff to given utxo iterator
func IteratorWithDiff(iterator model.ReadOnlyUTXOSetIterator, diff model.UTXODiff) (model.ReadOnlyUTXOSetIterator, error) {
	d, ok := diff.(*immutableUTXODiff)
	if !ok {
		return nil, errors.New("diff is not of type *immutableUTXODiff")
	}

	if iteratorWithDiff, ok := iterator.(*readOnlyUTXOIteratorWithDiff); ok {
		combinedDiff, err := iteratorWithDiff.diff.WithDiff(d)
		if err != nil {
			return nil, err
		}

		return IteratorWithDiff(iteratorWithDiff.baseIterator, combinedDiff)
	}

	return &readOnlyUTXOIteratorWithDiff{
		baseIterator:  iterator,
		diff:          d,
		toAddIterator: d.ToAdd().Iterator(),
	}, nil
}

func (r *readOnlyUTXOIteratorWithDiff) First() bool {
	baseNotEmpty := r.baseIterator.First()
	baseEmpty := !baseNotEmpty

	r.toAddIterator = r.diff.ToAdd().Iterator()
	toAddEmpty := r.diff.ToAdd().Len() == 0

	if baseEmpty {
		if toAddEmpty {
			return false
		}
		return r.Next()
	}

	r.currentOutpoint, r.currentUTXOEntry, r.currentErr = r.baseIterator.Get()
	if r.diff.mutableUTXODiff.toRemove.containsWithBlueScore(r.currentOutpoint, r.currentUTXOEntry.BlockBlueScore()) {
		return r.Next()
	}
	return true
}

func (r *readOnlyUTXOIteratorWithDiff) Next() bool {
	for r.baseIterator.Next() { // keep looping until we reach an outpoint/entry pair that is not in r.diff.toRemove
		r.currentOutpoint, r.currentUTXOEntry, r.currentErr = r.baseIterator.Get()
		if !r.diff.mutableUTXODiff.toRemove.containsWithBlueScore(r.currentOutpoint, r.currentUTXOEntry.BlockBlueScore()) {
			return true
		}
	}

	if r.toAddIterator.Next() {
		r.currentOutpoint, r.currentUTXOEntry, r.currentErr = r.toAddIterator.Get()
		return true
	}

	return false
}

func (r *readOnlyUTXOIteratorWithDiff) Get() (outpoint *externalapi.DomainOutpoint, utxoEntry externalapi.UTXOEntry, err error) {
	return r.currentOutpoint, r.currentUTXOEntry, r.currentErr
}
