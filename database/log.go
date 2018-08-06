// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package database

import (
	"github.com/btcsuite/btclog"
	"github.com/daglabs/btcd/logger"
)

// log is a logger that is initialized with no output filters.  This
// means the package will not perform any logging by default until the caller
// requests it.
var log btclog.Logger

// The default amount of logging is none.
func init() {
	log, _ = logger.Get(logger.SubsystemTags.BCDB)
	// Update the logger for the registered drivers.
	for _, drv := range drivers {
		if drv.UseLogger != nil {
			drv.UseLogger(log)
		}
	}
}
