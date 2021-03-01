package app

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/infrastructure/os/execenv"
	"github.com/kaspanet/kaspad/infrastructure/os/limits"
	"github.com/kaspanet/kaspad/infrastructure/os/signal"
	"github.com/kaspanet/kaspad/infrastructure/os/winservice"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/kaspanet/kaspad/util/profiling"
	"github.com/kaspanet/kaspad/version"
)

const leveldbCacheSizeMiB = 256

var desiredLimits = &limits.DesiredLimits{
	FileLimitWant: 2048,
	FileLimitMin:  1024,
}

var serviceDescription = &winservice.ServiceDescription{
	Name:        "kaspadsvc",
	DisplayName: "Kaspad Service",
	Description: "Downloads and stays synchronized with the Kaspa blockDAG and " +
		"provides DAG services to applications.",
}

type kaspadApp struct {
	cfg *config.Config
}

// StartApp starts the kaspad app, and blocks until it finishes running
func StartApp() error {
	execenv.Initialize(desiredLimits)

	// Load configuration and parse command line. This function also
	// initializes logging and configures it accordingly.
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	defer logger.BackendLog.Close()
	defer panics.HandlePanic(log, "MAIN", nil)

	app := &kaspadApp{cfg: cfg}

	// Call serviceMain on Windows to handle running as a service. When
	// the return isService flag is true, exit now since we ran as a
	// service. Otherwise, just fall through to normal operation.
	if runtime.GOOS == "windows" {
		isService, err := winservice.WinServiceMain(app.main, serviceDescription, cfg)
		if err != nil {
			return err
		}
		if isService {
			return nil
		}
	}

	return app.main(nil)
}

func (app *kaspadApp) main(startedChan chan<- struct{}) error {
	// Get a channel that will be closed when a shutdown signal has been
	// triggered either from an OS signal such as SIGINT (Ctrl+C) or from
	// another subsystem such as the RPC server.
	interrupt := signal.InterruptListener()
	defer log.Info("Shutdown complete")

	// Show version at startup.
	log.Infof("Version %s", version.Version())

	// Enable http profiling server if requested.
	if app.cfg.Profile != "" {
		profiling.Start(app.cfg.Profile, log)
	}

	// Perform upgrades to kaspad as new versions require it.
	if err := doUpgrades(); err != nil {
		log.Error(err)
		return err
	}

	// Return now if an interrupt signal was triggered.
	if signal.InterruptRequested(interrupt) {
		return nil
	}

	if app.cfg.ResetDatabase {
		err := removeDatabase(app.cfg)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	// Open the database
	databaseContext, err := openDB(app.cfg)
	if err != nil {
		log.Error(err)
		return err
	}

	defer func() {
		log.Infof("Gracefully shutting down the database...")
		err := databaseContext.Close()
		if err != nil {
			log.Errorf("Failed to close the database: %s", err)
		}
	}()

	// Return now if an interrupt signal was triggered.
	if signal.InterruptRequested(interrupt) {
		return nil
	}

	// Create componentManager and start it.
	componentManager, err := NewComponentManager(app.cfg, databaseContext, interrupt)
	if err != nil {
		log.Errorf("Unable to start kaspad: %+v", err)
		return err
	}

	defer func() {
		log.Infof("Gracefully shutting down kaspad...")

		shutdownDone := make(chan struct{})
		go func() {
			componentManager.Stop()
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

	componentManager.Start()

	if startedChan != nil {
		startedChan <- struct{}{}
	}

	// Wait until the interrupt signal is received from an OS signal or
	// shutdown is requested through one of the subsystems such as the RPC
	// server.
	<-interrupt
	return nil
}

// doUpgrades performs upgrades to kaspad as new versions require it.
// currently it's a placeholder we got from kaspad upstream, that does nothing
func doUpgrades() error {
	return nil
}

// dbPath returns the path to the block database given a database type.
func databasePath(cfg *config.Config) string {
	return filepath.Join(cfg.DataDir, "db")
}

func removeDatabase(cfg *config.Config) error {
	dbPath := databasePath(cfg)
	return os.RemoveAll(dbPath)
}

func openDB(cfg *config.Config) (database.Database, error) {
	dbPath := databasePath(cfg)
	log.Infof("Loading database from '%s'", dbPath)
	return ldb.NewLevelDB(dbPath, leveldbCacheSizeMiB)
}
