// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package server

import (
	"github.com/daglabs/btcd/logger"
	"github.com/daglabs/btcd/logs"
	"github.com/daglabs/btcd/util/panics"
)

// log is a logger that is initialized with no output filters.  This
// means the package will not perform any logging by default until the caller
// requests it.
var srvrLog, peerLog, txmpLog, indxLog, rpcsLog, amgrLog logs.Logger
var spawn func(func())

func init() {
	srvrLog, _ = logger.Get(logger.SubsystemTags.SRVR)
	peerLog, _ = logger.Get(logger.SubsystemTags.PEER)
	txmpLog, _ = logger.Get(logger.SubsystemTags.TXMP)
	indxLog, _ = logger.Get(logger.SubsystemTags.INDX)
	rpcsLog, _ = logger.Get(logger.SubsystemTags.RPCS)
	amgrLog, _ = logger.Get(logger.SubsystemTags.AMGR)

	spawn = panics.GoroutineWrapperFunc(srvrLog, logger.BackendLog)
}
