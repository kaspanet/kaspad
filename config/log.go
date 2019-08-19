// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package config

import (
	"github.com/daglabs/btcd/logger"
	"github.com/daglabs/btcd/logs"
)

// log is a logger that is initialized with no output filters.  This
// means the package will not perform any logging by default until the caller
// requests it.
var log logs.Logger

func init() {
	log, _ = logger.Get(logger.SubsystemTags.CNFG)
}
