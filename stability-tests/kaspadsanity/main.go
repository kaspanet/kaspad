package main

import (
	"fmt"
	"os"

	"github.com/zoomy-network/zoomyd/stability-tests/common"
	"github.com/zoomy-network/zoomyd/util/profiling"

	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/util/panics"
)

func main() {
	defer panics.HandlePanic(log, "kaspadsanity-main", nil)
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

	argsChan := readArgs()
	failures, err := commandLoop(argsChan)
	if err != nil {
		panic(errors.Wrap(err, "error in commandLoop"))
	}

	if len(failures) > 0 {
		fmt.Fprintf(os.Stderr, "FAILED:\n")
		for _, failure := range failures {
			fmt.Fprintln(os.Stderr, failure)
		}
		backendLog.Close()
		os.Exit(1)
	}

	log.Infof("All tests have passed")
}
