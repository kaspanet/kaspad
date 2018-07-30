package blockdag

import (
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/wire"
)

// utxoIteratorOutput represents all fields of a single UTXO, to be returned by an iterator
type utxoIteratorOutput struct {
	hash  daghash.Hash
	index uint32
	txOut *wire.TxOut
}

// utxoIterator is used to iterate over a utxoSet
type utxoIterator <-chan utxoIteratorOutput

// iterate returns an iterator for a UTXOCollection, and therefore, also a fullUTXOSet
func (c utxoCollection) iterate() utxoIterator {
	iterator := make(chan utxoIteratorOutput)

	go func() {
		for hash, txOuts := range c {
			for index, txOut := range txOuts {
				iterator <- utxoIteratorOutput{
					hash:  hash,
					index: index,
					txOut: txOut,
				}
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
		for utxo := range u.base.iterate() {
			if !u.utxoDiff.toRemove.contains(utxo.hash, utxo.index) {
				iterator <- utxo
			}
		}

		for utxo := range u.utxoDiff.toAdd.iterate() {
			iterator <- utxo
		}
		close(iterator)
	}()

	return iterator
}
