// Copyright (c) 2015-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package peer

import (
	"github.com/kaspanet/kaspad/logger"
	"github.com/kaspanet/kaspad/util/panics"
)

var log, _ = logger.Get(logger.SubsystemTags.PEER)
var spawn = panics.GoroutineWrapperFunc(log)
var spawnAfter = panics.AfterFuncWrapperFunc(log)
