// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package netsync

import (
	"github.com/daglabs/kaspad/logger"
	"github.com/daglabs/kaspad/util/panics"
)

var log, _ = logger.Get(logger.SubsystemTags.SYNC)
var spawn = panics.GoroutineWrapperFuncWithPanicHandler(log)
