// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"runtime"

	"github.com/kaspanet/kaspad/limits"
	"github.com/kaspanet/kaspad/logs"
	"github.com/kaspanet/kaspad/util/panics"
)

const (
	// blockDBNamePrefix is the prefix for the kaspad block database.
	blockDBNamePrefix = "blocks"
)

var (
	cfg   *ConfigFlags
	log   *logs.Logger
	spawn func(func())
)

// realMain is the real main function for the utility. It is necessary to work
// around the fact that deferred functions do not run when os.Exit() is called.
func realMain() error {
	// Load configuration and parse command line.
	tcfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	cfg = tcfg

	// Setup logging.
	backendLogger := logs.NewBackend()
	defer os.Stdout.Sync()
	log = backendLogger.Logger("MAIN")
	spawn = panics.GoroutineWrapperFunc(log)

	fi, err := os.Open(cfg.InFile)
	if err != nil {
		log.Errorf("Failed to open file %s: %s", cfg.InFile, err)
		return err
	}
	defer fi.Close()

	// Create a block importer for the database and input file and start it.
	// The done channel returned from start will contain an error if
	// anything went wrong.
	importer, err := newBlockImporter(fi)
	if err != nil {
		log.Errorf("Failed create block importer: %s", err)
		return err
	}

	// Perform the import asynchronously. This allows blocks to be
	// processed and read in parallel. The results channel returned from
	// Import contains the statistics about the import including an error
	// if something went wrong.
	log.Info("Starting import")
	resultsChan := importer.Import()
	results := <-resultsChan
	if results.err != nil {
		log.Errorf("%s", results.err)
		return results.err
	}

	log.Infof("Processed a total of %d blocks (%d imported, %d already "+
		"known)", results.blocksProcessed, results.blocksImported,
		results.blocksProcessed-results.blocksImported)
	return nil
}

func main() {
	// Use all processor cores and up some limits.
	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := limits.SetLimits(); err != nil {
		os.Exit(1)
	}

	// Work around defer not working after os.Exit()
	if err := realMain(); err != nil {
		os.Exit(1)
	}
}
