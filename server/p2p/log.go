// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package p2p

import (
	"github.com/kaspanet/kaspad/logger"
	"github.com/kaspanet/kaspad/util/panics"
)

var (
	srvrLog, _ = logger.Get(logger.SubsystemTags.SRVR)
	peerLog, _ = logger.Get(logger.SubsystemTags.PEER)
	spawn      = panics.GoroutineWrapperFunc(logger.BackendLog)

	txmpLog, _ = logger.Get(logger.SubsystemTags.TXMP)
	indxLog, _ = logger.Get(logger.SubsystemTags.INDX)
	amgrLog, _ = logger.Get(logger.SubsystemTags.AMGR)
)
