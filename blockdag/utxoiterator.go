package blockdag

import (
	"github.com/daglabs/btcd/wire"
)

// utxoIteratorOutput represents all fields of a single UTXO, to be returned by an iterator
type utxoIteratorOutput struct {
	outPoint wire.OutPoint
	entry    *UtxoEntry
}

// utxoIterator is used to iterate over a utxoSet
type utxoIterator <-chan utxoIteratorOutput

type utxoIterable interface {
	iterate() utxoIterator
}

// iterate returns an iterator for a UTXOCollection, and therefore, also a fullUTXOSet
func (c utxoCollection) iterate() utxoIterator {
	iterator := make(chan utxoIteratorOutput)

	go func() {
		for outPoint, entry := range c {
			iterator <- utxoIteratorOutput{
				outPoint: outPoint,
				entry:    entry,
			}
		}
		close(iterator)
	}()

	return iterator
}

// iterate returns an iterator for a diffUTXOSet
func (u *diffUTXOSet) iterate() utxoIterator {
	iterator := make(chan utxoIteratorOutput)

	go func() {
		for output := range u.base.iterate() {
			if _, ok := u.utxoDiff.toRemove[output.outPoint]; !ok {
				iterator <- output
			}
		}

		for utxo := range u.utxoDiff.toAdd.iterate() {
			iterator <- utxo
		}
		close(iterator)
	}()

	return iterator
}
