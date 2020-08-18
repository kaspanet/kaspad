// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package client

import (
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/panics"
)

// log is a logger that is initialized with no output filters. This
// means the package will not perform any logging by default until the caller
// requests it.
var log *logger.Logger
var spawn func(name string, f func())

const logSubsytem = "RPCC"

// The default amount of logging is none.
func init() {
	DisableLog()
}

// DisableLog disables all library log output. Logging output is disabled
// by default until UseLogger is called.
func DisableLog() {
	backend := logger.NewBackend()
	log = backend.Logger(logSubsytem)
	log.SetLevel(logger.LevelOff)
	spawn = panics.GoroutineWrapperFunc(log)
}

// UseLogger uses a specified Logger to output package logging info.
func UseLogger(backend *logger.Backend, level logger.Level) {
	log = backend.Logger(logSubsytem)
	log.SetLevel(level)
	spawn = panics.GoroutineWrapperFunc(log)
}
