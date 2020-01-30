// Copyright (c) 2015-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/logger"
	"github.com/kaspanet/kaspad/logs"
)

const (
	// blockDbNamePrefix is the prefix for the kaspad block database.
	blockDbNamePrefix = "blocks"
)

var (
	log             logs.Logger
	spawn           func(func())
	shutdownChannel = make(chan error)
)

// loadBlockDB opens the block database and returns a handle to it.
func loadBlockDB() (database.DB, error) {
	// The database name is based on the database type.
	dbName := blockDbNamePrefix + "_" + cfg.DbType
	dbPath := filepath.Join(cfg.DataDir, dbName)

	log.Infof("Loading block database from '%s'", dbPath)
	db, err := database.Open(cfg.DbType, dbPath, activeNetParams.Net)
	if err != nil {
		// Return the error if it's not because the database doesn't
		// exist.
		var dbErr database.Error
		if ok := errors.As(err, &dbErr); !ok || dbErr.ErrorCode !=
			database.ErrDbDoesNotExist {

			return nil, err
		}

		// Create the db if it does not exist.
		err = os.MkdirAll(cfg.DataDir, 0700)
		if err != nil {
			return nil, err
		}
		db, err = database.Create(cfg.DbType, dbPath, activeNetParams.Net)
		if err != nil {
			return nil, err
		}
	}

	log.Info("Block database loaded")
	return db, nil
}

// realMain is the real main function for the utility. It is necessary to work
// around the fact that deferred functions do not run when os.Exit() is called.
func realMain() error {
	// Setup logging.
	backendLogger := logs.NewBackend()
	defer os.Stdout.Sync()
	log = backendLogger.Logger("MAIN")
	spawn = panics.GoroutineWrapperFunc(log)
	dbLog, _ := logger.Get(logger.SubsystemTags.BCDB)
	dbLog.SetLevel(logs.LevelDebug)

	// Setup the parser options and commands.
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	parserFlags := flags.Options(flags.HelpFlag | flags.PassDoubleDash)
	parser := flags.NewNamedParser(appName, parserFlags)
	parser.AddGroup("Global Options", "", cfg)
	parser.AddCommand("insecureimport",
		"Insecurely import bulk block data from bootstrap.dat",
		"Insecurely import bulk block data from bootstrap.dat. "+
			"WARNING: This is NOT secure because it does NOT "+
			"verify DAG rules. It is only provided for testing "+
			"purposes.", &importCfg)
	parser.AddCommand("loadheaders",
		"Time how long to load headers for all blocks in the database",
		"", &headersCfg)
	parser.AddCommand("fetchblock",
		"Fetch the specific block hash from the database", "",
		&fetchBlockCfg)
	parser.AddCommand("fetchblockregion",
		"Fetch the specified block region from the database", "",
		&blockRegionCfg)

	// Parse command line and invoke the Execute function for the specified
	// command.
	if _, err := parser.Parse(); err != nil {
		var e *flags.Error
		if ok := errors.As(err, &e); ok && e.Type == flags.ErrHelp {
			parser.WriteHelp(os.Stderr)
		} else {
			log.Error(err)
		}

		return err
	}

	return nil
}

func main() {
	// Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Work around defer not working after os.Exit()
	if err := realMain(); err != nil {
		os.Exit(1)
	}
}
