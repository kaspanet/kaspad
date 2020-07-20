// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/kaspanet/kaspad/dbaccess"

	"github.com/kaspanet/kaspad/blockdag/indexers"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/limits"
	"github.com/kaspanet/kaspad/signal"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/kaspanet/kaspad/util/profiling"
	"github.com/kaspanet/kaspad/version"
)

const (
	// blockDbNamePrefix is the prefix for the block database name. The
	// database type is appended to this value to form the full block
	// database name.
	blockDbNamePrefix = "blocks"
)

// winServiceMain is only invoked on Windows. It detects when kaspad is running
// as a service and reacts accordingly.
var winServiceMain func() (bool, error)

// kaspadMain is the real main function for kaspad. It is necessary to work
// around the fact that deferred functions do not run when os.Exit() is called.
// The optional startedChan writes once all services has started.
func kaspadMain(startedChan chan<- struct{}) error {
	interrupt := signal.InterruptListener()

	// Load configuration and parse command line. This function also
	// initializes logging and configures it accordingly.
	err := config.LoadAndSetActiveConfig()
	if err != nil {
		return err
	}
	cfg := config.ActiveConfig()
	defer panics.HandlePanic(log, "MAIN", nil)

	// Get a channel that will be closed when a shutdown signal has been
	// triggered either from an OS signal such as SIGINT (Ctrl+C) or from
	// another subsystem such as the RPC server.
	defer log.Info("Shutdown complete")

	// Show version at startup.
	log.Infof("Version %s", version.Version())

	// Enable http profiling server if requested.
	if cfg.Profile != "" {
		profiling.Start(cfg.Profile, log)
	}

	// Write cpu profile if requested.
	if cfg.CPUProfile != "" {
		f, err := os.Create(cfg.CPUProfile)
		if err != nil {
			log.Errorf("Unable to create cpu profile: %s", err)
			return err
		}
		pprof.StartCPUProfile(f)
		defer f.Close()
		defer pprof.StopCPUProfile()
	}

	// Perform upgrades to kaspad as new versions require it.
	if err := doUpgrades(); err != nil {
		log.Errorf("%s", err)
		return err
	}

	// Return now if an interrupt signal was triggered.
	if signal.InterruptRequested(interrupt) {
		return nil
	}

	if cfg.ResetDatabase {
		err := removeDatabase()
		if err != nil {
			log.Errorf("%s", err)
			return err
		}
	}

	// Open the database
	err = openDB()
	if err != nil {
		log.Errorf("%s", err)
		return err
	}
	defer func() {
		log.Infof("Gracefully shutting down the database...")
		err := dbaccess.Close()
		if err != nil {
			log.Errorf("Failed to close the database: %s", err)
		}
	}()

	// Return now if an interrupt signal was triggered.
	if signal.InterruptRequested(interrupt) {
		return nil
	}

	// Drop indexes and exit if requested.
	if cfg.DropAcceptanceIndex {
		if err := indexers.DropAcceptanceIndex(); err != nil {
			log.Errorf("%s", err)
			return err
		}

		return nil
	}

	// Create kaspad and start it.
	kaspad, err := newKaspad(interrupt)
	if err != nil {
		log.Errorf("Unable to start kaspad: %+v", err)
		return err
	}
	defer func() {
		log.Infof("Gracefully shutting down kaspad...")
		kaspad.stop()

		shutdownDone := make(chan struct{})
		go func() {
			kaspad.WaitForShutdown()
			shutdownDone <- struct{}{}
		}()

		const shutdownTimeout = 2 * time.Minute

		select {
		case <-shutdownDone:
		case <-time.After(shutdownTimeout):
			log.Criticalf("Graceful shutdown timed out %s. Terminating...", shutdownTimeout)
		}
		log.Infof("Kaspad shutdown complete")
	}()
	kaspad.start()
	if startedChan != nil {
		startedChan <- struct{}{}
	}

	// Wait until the interrupt signal is received from an OS signal or
	// shutdown is requested through one of the subsystems such as the RPC
	// server.
	<-interrupt
	return nil
}

func removeDatabase() error {
	dbPath := blockDbPath(config.ActiveConfig().DbType)
	return os.RemoveAll(dbPath)
}

// removeRegressionDB removes the existing regression test database if running
// in regression test mode and it already exists.

// dbPath returns the path to the block database given a database type.
func blockDbPath(dbType string) string {
	// The database name is based on the database type.
	dbName := blockDbNamePrefix + "_" + dbType
	if dbType == "sqlite" {
		dbName = dbName + ".db"
	}
	dbPath := filepath.Join(config.ActiveConfig().DataDir, dbName)
	return dbPath
}

func openDB() error {
	dbPath := filepath.Join(config.ActiveConfig().DataDir, "db")
	log.Infof("Loading database from '%s'", dbPath)
	err := dbaccess.Open(dbPath)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	// Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Up some limits.
	if err := limits.SetLimits(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set limits: %s\n", err)
		os.Exit(1)
	}

	// Call serviceMain on Windows to handle running as a service. When
	// the return isService flag is true, exit now since we ran as a
	// service. Otherwise, just fall through to normal operation.
	if runtime.GOOS == "windows" {
		isService, err := winServiceMain()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if isService {
			os.Exit(0)
		}
	}

	// Work around defer not working after os.Exit()
	if err := kaspadMain(nil); err != nil {
		os.Exit(1)
	}
}
