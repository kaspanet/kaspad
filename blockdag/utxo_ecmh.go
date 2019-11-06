package blockdag

import (
	"bytes"
	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
	"github.com/golang/groupcache/lru"
	"sync"
)

const ecmhCacheSize = 4_000_000

var (
	ecmhCache     = lru.New(ecmhCacheSize)
	ecmhCacheLock sync.Mutex
)

func utxoMultiset(entry *UTXOEntry, outpoint *wire.Outpoint) (*btcec.Multiset, error) {
	w := &bytes.Buffer{}
	err := serializeUTXO(w, entry, outpoint)
	if err != nil {
		return nil, err
	}
	serializedUTXO := w.Bytes()
	utxoHash := daghash.DoubleHashH(serializedUTXO)

	ecmhCacheLock.Lock()
	defer ecmhCacheLock.Unlock()

	if cachedMSPoint, ok := ecmhCache.Get(utxoHash); ok {
		return cachedMSPoint.(*btcec.Multiset), nil
	}
	msPoint := btcec.NewMultiset(btcec.S256()).Add(serializedUTXO)
	ecmhCache.Add(utxoHash, msPoint)
	return msPoint, nil
}
