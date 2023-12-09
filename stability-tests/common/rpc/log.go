package rpc

import (
	"github.com/zoomy-network/zoomyd/infrastructure/logger"
	"github.com/zoomy-network/zoomyd/util/panics"
)

// log is a logger that is initialized with no output filters. This
// means the package will not perform any logging by default until the caller
// requests it.
var log *logger.Logger
var spawn func(name string, spawnedFunction func())

const logSubsytem = "CRPC"

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
