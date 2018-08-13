// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"github.com/daglabs/btcd/logger"
)

var btcdLog, _ = logger.Get(logger.SubsystemTags.BTCD)
var srvrLog, _ = logger.Get(logger.SubsystemTags.SRVR)
