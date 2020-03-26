package blockdag

import (
	"bytes"
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/wire"
)

func addUTXOToMultiset(ms *secp256k1.MultiSet, entry *UTXOEntry, outpoint *wire.Outpoint) (*secp256k1.MultiSet, error) {
	w := &bytes.Buffer{}
	err := serializeUTXO(w, entry, outpoint)
	if err != nil {
		return nil, err
	}
	ms.Add(w.Bytes())
	return ms, nil
}

func removeUTXOFromMultiset(ms *secp256k1.MultiSet, entry *UTXOEntry, outpoint *wire.Outpoint) (*secp256k1.MultiSet, error) {
	w := &bytes.Buffer{}
	err := serializeUTXO(w, entry, outpoint)
	if err != nil {
		return nil, err
	}
	ms.Remove(w.Bytes())
	return ms, nil
}
