package bigintpool

import (
	"math/big"
	"sync"
)

var bigIntPool = sync.Pool{
	New: func() interface{} {
		return big.NewInt(0)
	},
}

// Acquire acquires a big.Int from the pool and
// initializes it to x.
func Acquire(x int64) *big.Int {
	bigInt := bigIntPool.Get().(*big.Int)
	bigInt.SetInt64(x)
	return bigInt
}

// Release returns the given big.Int to the pool.
func Release(toRelease *big.Int) {
	toRelease.SetInt64(0)
	bigIntPool.Put(toRelease)
}
