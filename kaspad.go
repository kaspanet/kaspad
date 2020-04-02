// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/dbaccess"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strings"

	"github.com/kaspanet/kaspad/blockdag/indexers"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/limits"
	"github.com/kaspanet/kaspad/server"
	"github.com/kaspanet/kaspad/signal"
	"github.com/kaspanet/kaspad/util/fs"
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

var (
	cfg *config.Config
)

// winServiceMain is only invoked on Windows. It detects when kaspad is running
// as a service and reacts accordingly.
var winServiceMain func() (bool, error)

// kaspadMain is the real main function for kaspad. It is necessary to work
// around the fact that deferred functions do not run when os.Exit() is called.
// The optional serverChan parameter is mainly used by the service code to be
// notified with the server once it is setup so it can gracefully stop it when
// requested from the service control manager.
func kaspadMain(serverChan chan<- *server.Server) error {
	interrupt := signal.InterruptListener()

	// Load configuration and parse command line. This function also
	// initializes logging and configures it accordingly.
	err := config.LoadAndSetActiveConfig()
	if err != nil {
		return err
	}
	cfg = config.ActiveConfig()
	defer panics.HandlePanic(kasdLog, nil)

	// Get a channel that will be closed when a shutdown signal has been
	// triggered either from an OS signal such as SIGINT (Ctrl+C) or from
	// another subsystem such as the RPC server.
	defer kasdLog.Info("Shutdown complete")

	// Show version at startup.
	kasdLog.Infof("Version %s", version.Version())

	// Enable http profiling server if requested.
	if cfg.Profile != "" {
		spawn(func() {
			profiling.Start(cfg.Profile, kasdLog)
		})
	}

	// Write cpu profile if requested.
	if cfg.CPUProfile != "" {
		f, err := os.Create(cfg.CPUProfile)
		if err != nil {
			kasdLog.Errorf("Unable to create cpu profile: %s", err)
			return err
		}
		pprof.StartCPUProfile(f)
		defer f.Close()
		defer pprof.StopCPUProfile()
	}

	// Perform upgrades to kaspad as new versions require it.
	if err := doUpgrades(); err != nil {
		kasdLog.Errorf("%s", err)
		return err
	}

	// Return now if an interrupt signal was triggered.
	if signal.InterruptRequested(interrupt) {
		return nil
	}

	if cfg.ResetDatabase {
		err := removeDatabase()
		if err != nil {
			kasdLog.Errorf("%s", err)
			return err
		}
	}

	// Open the database
	err = openDB()
	if err != nil {
		kasdLog.Errorf("%s", err)
		return err
	}
	defer func() {
		kasdLog.Infof("Gracefully shutting down the database...")
		err := dbaccess.Close()
		if err != nil {
			kasdLog.Errorf("Failed to close the database: %s", err)
		}
	}()

	// Return now if an interrupt signal was triggered.
	if signal.InterruptRequested(interrupt) {
		return nil
	}

	// Drop indexes and exit if requested.
	if cfg.DropAcceptanceIndex {
		if err := indexers.DropAcceptanceIndex(); err != nil {
			kasdLog.Errorf("%s", err)
			return err
		}

		return nil
	}

	// Create server and start it.
	server, err := server.NewServer(cfg.Listeners, config.ActiveConfig().NetParams(),
		interrupt)
	if err != nil {
		// TODO: this logging could do with some beautifying.
		kasdLog.Errorf("Unable to start server on %s: %s",
			strings.Join(cfg.Listeners, ", "), err)
		return err
	}
	defer func() {
		kasdLog.Infof("Gracefully shutting down the server...")
		server.Stop()
		server.WaitForShutdown()
		srvrLog.Infof("Server shutdown complete")
	}()
	server.Start()
	if serverChan != nil {
		serverChan <- server
	}

	// Wait until the interrupt signal is received from an OS signal or
	// shutdown is requested through one of the subsystems such as the RPC
	// server.
	<-interrupt
	return nil
}

func removeDatabase() error {
	dbPath := blockDbPath(cfg.DbType)
	return os.RemoveAll(dbPath)
}

// removeRegressionDB removes the existing regression test database if running
// in regression test mode and it already exists.
func removeRegressionDB(dbPath string) error {
	// Don't do anything if not in regression test mode.
	if !cfg.RegressionTest {
		return nil
	}

	// Remove the old regression test database if it already exists.
	fi, err := os.Stat(dbPath)
	if err == nil {
		kasdLog.Infof("Removing regression test database from '%s'", dbPath)
		if fi.IsDir() {
			err := os.RemoveAll(dbPath)
			if err != nil {
				return err
			}
		} else {
			err := os.Remove(dbPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// dbPath returns the path to the block database given a database type.
func blockDbPath(dbType string) string {
	// The database name is based on the database type.
	dbName := blockDbNamePrefix + "_" + dbType
	if dbType == "sqlite" {
		dbName = dbName + ".db"
	}
	dbPath := filepath.Join(cfg.DataDir, dbName)
	return dbPath
}

// warnMultipleDBs shows a warning if multiple block database types are detected.
// This is not a situation most users want. It is handy for development however
// to support multiple side-by-side databases.
func warnMultipleDBs() {
	// This is intentionally not using the known db types which depend
	// on the database types compiled into the binary since we want to
	// detect legacy db types as well.
	dbTypes := []string{"ffldb", "leveldb", "sqlite"}
	duplicateDbPaths := make([]string, 0, len(dbTypes)-1)
	for _, dbType := range dbTypes {
		if dbType == cfg.DbType {
			continue
		}

		// Store db path as a duplicate db if it exists.
		dbPath := blockDbPath(dbType)
		if fs.FileExists(dbPath) {
			duplicateDbPaths = append(duplicateDbPaths, dbPath)
		}
	}

	// Warn if there are extra databases.
	if len(duplicateDbPaths) > 0 {
		selectedDbPath := blockDbPath(cfg.DbType)
		kasdLog.Warnf("WARNING: There are multiple block DAG databases "+
			"using different database types.\nYou probably don't "+
			"want to waste disk space by having more than one.\n"+
			"Your current database is located at [%s].\nThe "+
			"additional database is located at %s", selectedDbPath,
			strings.Join(duplicateDbPaths, ", "))
	}
}

func openDB() error {
	dbPath := filepath.Join(cfg.DataDir, "db")
	kasdLog.Infof("Loading database from '%s'", dbPath)
	err := dbaccess.Open(dbPath)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	// Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Block and transaction processing can cause bursty allocations. This
	// limits the garbage collector from excessively overallocating during
	// bursts. This value was arrived at with the help of profiling live
	// usage.
	debug.SetGCPercent(10)

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
