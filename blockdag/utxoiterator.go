package blockdag

import (
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/wire"
)

// utxoIteratorOutput represents all fields of a single UTXO, to be returned by an iterator
type utxoIteratorOutput struct {
	previousHash daghash.Hash
	index        int
	txOut        *wire.TxOut
}

// utxoIterator is used to iterate over a UTXOSet
type utxoIterator <-chan utxoIteratorOutput

// Iterate returns an iterator for a UTXOCollection, and therefore, also a FullUTXOSet
func (c utxoCollection) Iterate() utxoIterator {
	iterator := make(chan utxoIteratorOutput)

	go func() {
		for previousHash, txOuts := range c {
			for index, txOut := range txOuts {
				iterator <- utxoIteratorOutput{
					previousHash: previousHash,
					index:        index,
					txOut:        txOut,
				}
			}
		}
		close(iterator)
	}()

	return iterator
}
//
//// Iterate returns an iterator for a DiffUTXOSet
//func (u *DiffUTXOSet) Iterate() utxoIterator {
//	iterator := make(chan utxoIteratorOutput)
//
//	go func() {
//		for utxo := range u.base.Iterate() {
//			if !u.diff.toRemove.Contains(utxo.PreviousID, utxo.index) {
//				iterator <- utxo
//			}
//		}
//
//		for utxo := range u.diff.toAdd.Iterate() {
//			iterator <- utxo
//		}
//		close(iterator)
//	}()
//
//	return iterator
//}
