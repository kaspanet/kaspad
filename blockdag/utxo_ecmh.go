package blockdag

import (
	"bytes"
	"github.com/golang/groupcache/lru"
	"github.com/kaspanet/kaspad/btcec"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

const ecmhCacheSize = 4_000_000

var (
	utxoToECMHCache = lru.New(ecmhCacheSize)
)

func utxoMultiset(entry *UTXOEntry, outpoint *wire.Outpoint) (*btcec.Multiset, error) {
	w := &bytes.Buffer{}
	err := serializeUTXO(w, entry, outpoint)
	if err != nil {
		return nil, err
	}
	serializedUTXO := w.Bytes()
	utxoHash := daghash.DoubleHashH(serializedUTXO)

	if cachedMSPoint, ok := utxoToECMHCache.Get(utxoHash); ok {
		return cachedMSPoint.(*btcec.Multiset), nil
	}
	msPoint := btcec.NewMultiset(btcec.S256()).Add(serializedUTXO)
	utxoToECMHCache.Add(utxoHash, msPoint)
	return msPoint, nil
}
