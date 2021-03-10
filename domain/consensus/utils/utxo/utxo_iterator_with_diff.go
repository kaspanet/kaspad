package utxo

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

type readOnlyUTXOIteratorWithDiff struct {
	baseIterator externalapi.ReadOnlyUTXOSetIterator
	diff         *immutableUTXODiff

	currentOutpoint  *externalapi.DomainOutpoint
	currentUTXOEntry externalapi.UTXOEntry
	currentErr       error

	toAddIterator externalapi.ReadOnlyUTXOSetIterator
	isClosed      bool
}

// IteratorWithDiff applies a UTXODiff to given utxo iterator
func IteratorWithDiff(iterator externalapi.ReadOnlyUTXOSetIterator, diff externalapi.UTXODiff) (externalapi.ReadOnlyUTXOSetIterator, error) {
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
	if r.isClosed {
		panic("Tried using a closed readOnlyUTXOIteratorWithDiff")
	}
	baseNotEmpty := r.baseIterator.First()
	baseEmpty := !baseNotEmpty

	err := r.toAddIterator.Close()
	if err != nil {
		r.currentErr = err
		return true
	}
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
	if r.isClosed {
		panic("Tried using a closed readOnlyUTXOIteratorWithDiff")
	}
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
	if r.isClosed {
		return nil, nil, errors.New("Tried using a closed readOnlyUTXOIteratorWithDiff")
	}
	return r.currentOutpoint, r.currentUTXOEntry, r.currentErr
}

func (r *readOnlyUTXOIteratorWithDiff) Close() error {
	if r.isClosed {
		return errors.New("Tried using a closed readOnlyUTXOIteratorWithDiff")
	}
	r.isClosed = true
	err := r.baseIterator.Close()
	if err != nil {
		return err
	}
	err = r.toAddIterator.Close()
	if err != nil {
		return err
	}
	r.baseIterator = nil
	r.diff = nil
	r.currentOutpoint = nil
	r.currentUTXOEntry = nil
	r.currentErr = nil
	r.toAddIterator = nil
	return nil
}
