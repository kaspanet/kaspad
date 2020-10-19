package blockdag

import (
	"bytes"

	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/utxo"
)

func addUTXOToMultiset(ms *secp256k1.MultiSet, entry *utxo.Entry, outpoint *appmessage.Outpoint) (*secp256k1.MultiSet, error) {
	w := &bytes.Buffer{}
	err := utxo.SerializeUTXO(w, entry, outpoint)
	if err != nil {
		return nil, err
	}
	ms.Add(w.Bytes())
	return ms, nil
}

func removeUTXOFromMultiset(ms *secp256k1.MultiSet, entry *utxo.Entry, outpoint *appmessage.Outpoint) (*secp256k1.MultiSet, error) {
	w := &bytes.Buffer{}
	err := utxo.SerializeUTXO(w, entry, outpoint)
	if err != nil {
		return nil, err
	}
	ms.Remove(w.Bytes())
	return ms, nil
}
