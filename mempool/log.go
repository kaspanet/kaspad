// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mempool

import (
	"github.com/btcsuite/btclog"
	"github.com/daglabs/btcd/logger"
)

// log is a logger that is initialized with no output filters.  This
// means the package will not perform any logging by default until the caller
// requests it.
var log btclog.Logger

func init() {
	log, _ = logger.Get(logger.SubsystemTags.TXMP)
}

// pickNoun returns the singular or plural form of a noun depending
// on the count n.
func pickNoun(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}
