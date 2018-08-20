// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package signal

import (
	"github.com/btcsuite/btclog"
	"github.com/daglabs/btcd/logger"
)

// log is a logger that is initialized with no output filters.  This
// means the package will not perform any logging by default until the caller
// requests it.
var btcdLog btclog.Logger

func init() {
	btcdLog, _ = logger.Get(logger.SubsystemTags.BTCD)
}
