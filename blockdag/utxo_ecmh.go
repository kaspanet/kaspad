package blockdag

import (
	"bytes"
	"github.com/golang/groupcache/lru"
	"github.com/kaspanet/kaspad/ecc"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

const ecmhCacheSize = 4_000_000

var (
	utxoToECMHCache = lru.New(ecmhCacheSize)
)

func utxoMultiset(entry *UTXOEntry, outpoint *wire.Outpoint) (*ecc.Multiset, error) {
	w := &bytes.Buffer{}
	err := serializeUTXO(w, entry, outpoint, false)
	if err != nil {
		return nil, err
	}
	serializedUTXO := w.Bytes()
	utxoHash := daghash.DoubleHashH(serializedUTXO)

	if cachedMSPoint, ok := utxoToECMHCache.Get(utxoHash); ok {
		return cachedMSPoint.(*ecc.Multiset), nil
	}
	msPoint := ecc.NewMultiset(ecc.S256()).Add(serializedUTXO)
	utxoToECMHCache.Add(utxoHash, msPoint)
	return msPoint, nil
}
