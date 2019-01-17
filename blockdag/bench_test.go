// Copyright (c) 2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"testing"

	"github.com/daglabs/btcd/util"
)

// BenchmarkIsCoinBase performs a simple benchmark against the IsCoinBase
// function.
func BenchmarkIsCoinBase(b *testing.B) {
	tx, _ := util.NewBlock(&Block100000).Tx(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsCoinBase(tx)
	}
}

// BenchmarkIsCoinBaseMethod performs a simple benchmark against the IsCoinBase
// method.
func BenchmarkIsCoinBaseMethod(b *testing.B) {
	tx := Block100000.Transactions[1]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx.IsCoinBase()
	}
}
