package blockdag

import (
	"bytes"
	"github.com/kaspanet/kaspad/ecc"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

func addUTXOToMultiset(ms *ecc.Multiset, entry *UTXOEntry, outpoint *wire.Outpoint) (*ecc.Multiset, error) {
	w := &bytes.Buffer{}
	err := serializeUTXO(w, entry, outpoint)
	if err != nil {
		return nil, err
	}
	serializedUTXO := w.Bytes()
	utxoHash := daghash.DoubleHashH(serializedUTXO)
	return ms.Add(utxoHash[:]), nil
}

func removeUTXOFromMultiset(ms *ecc.Multiset, entry *UTXOEntry, outpoint *wire.Outpoint) (*ecc.Multiset, error) {
	w := &bytes.Buffer{}
	err := serializeUTXO(w, entry, outpoint)
	if err != nil {
		return nil, err
	}
	serializedUTXO := w.Bytes()
	utxoHash := daghash.DoubleHashH(serializedUTXO)
	return ms.Remove(utxoHash[:]), nil
}
