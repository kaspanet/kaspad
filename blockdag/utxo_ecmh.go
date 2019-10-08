package blockdag

import (
	"bytes"
	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
	"github.com/golang/groupcache/lru"
)

const ecmhCacheSize = 1 >> 22

var ecmhCache = lru.New(ecmhCacheSize)

func utxoMultiset(entry *UTXOEntry, outpoint *wire.Outpoint) (*btcec.Multiset, error) {
	w := &bytes.Buffer{}
	err := serializeUTXO(w, entry, outpoint)
	if err != nil {
		return nil, err
	}
	serializedUTXO := w.Bytes()
	utxoHash := daghash.DoubleHashH(serializedUTXO)
	cachedMSPoint, ok := ecmhCache.Get(utxoHash)
	if ok {
		return cachedMSPoint.(*btcec.Multiset), nil
	}
	msPoint := btcec.NewMultiset(btcec.S256()).Add(serializedUTXO)
	ecmhCache.Add(utxoHash, msPoint)
	return msPoint, nil
}
