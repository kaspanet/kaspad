// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package p2p

import (
	"github.com/btcsuite/btclog"
	"github.com/daglabs/btcd/logger"
)

// log is a logger that is initialized with no output filters.  This
// means the package will not perform any logging by default until the caller
// requests it.
var srvLog, peerLog btclog.Logger

func init() {
	srvLog, _ = logger.Get(logger.SubsystemTags.SRVR)
	peerLog, _ = logger.Get(logger.SubsystemTags.PEER)
}
