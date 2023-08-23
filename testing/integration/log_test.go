// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package integration

import (
	"github.com/c4ei/kaspad/infrastructure/logger"
	"github.com/c4ei/kaspad/util/panics"
)

var log = logger.RegisterSubSystem("INTG")
var spawn = panics.GoroutineWrapperFunc(log)
