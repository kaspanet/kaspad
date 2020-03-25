package blockdag

import (
	"bytes"
	"github.com/kaspanet/kaspad/ecc"
	"github.com/kaspanet/kaspad/wire"
)

func addUTXOToMultiset(ms *ecc.Multiset, entry *UTXOEntry, outpoint *wire.Outpoint) (*ecc.Multiset, error) {
	w := &bytes.Buffer{}
	err := serializeUTXO(w, entry, outpoint)
	if err != nil {
		return nil, err
	}
	serializedUTXO := w.Bytes()
	return ms.Add(serializedUTXO), nil
}

func removeUTXOFromMultiset(ms *ecc.Multiset, entry *UTXOEntry, outpoint *wire.Outpoint) (*ecc.Multiset, error) {
	w := &bytes.Buffer{}
	err := serializeUTXO(w, entry, outpoint)
	if err != nil {
		return nil, err
	}
	serializedUTXO := w.Bytes()
	return ms.Remove(serializedUTXO), nil
}
