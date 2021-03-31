package main

import (
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/util/profiling"

	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
)

func main() {
	defer panics.HandlePanic(log, "mempool-limits-main", nil)
	err := parseConfig()
	if err != nil {
		panic(errors.Wrap(err, "error in parseConfig"))
	}
	defer backendLog.Close()
	common.UseLogger(backendLog, log.Level())

	cfg := activeConfig()
	if cfg.Profile != "" {
		profiling.Start(cfg.Profile, log)
	}

	log.Infof("All tests have passed")
}
