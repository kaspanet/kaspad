package main

import (
	"path/filepath"

	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/stability-tests/common"

	"github.com/jessevdk/go-flags"
)

const (
	defaultLogFilename    = "netsync.log"
	defaultErrLogFilename = "netsync_err.log"
)

var (
	// Default configuration options
	defaultLogFile    = filepath.Join(common.DefaultAppDir, defaultLogFilename)
	defaultErrLogFile = filepath.Join(common.DefaultAppDir, defaultErrLogFilename)
)

type configFlags struct {
	LogLevel            string `short:"d" long:"loglevel" description:"Set log level {trace, debug, info, warn, error, critical}"`
	DAGFile             string `long:"dag-file" description:"Path to DAG JSON file"`
	Profile             string `long:"profile" description:"Enable HTTP profiling on given port -- NOTE port must be between 1024 and 65536"`
	MiningDataDirectory string `long:"mining-data-dir" description:"Mining Data directory (will generate a random one if omitted)"`
	SyncerDataDirectory string `long:"syncer-data-dir" description:"Syncer Data directory (will generate a random one if omitted)"`
	SynceeDataDirectory string `long:"syncee-data-dir" description:"Syncee Data directory (will generate a random one if omitted)"`
	config.NetworkFlags
}

var cfg *configFlags

func activeConfig() *configFlags {
	return cfg
}

func parseConfig() error {
	cfg = &configFlags{}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()
	if err != nil {
		return err
	}

	err = cfg.ResolveNetwork(parser)
	if err != nil {
		return err
	}

	initLog(defaultLogFile, defaultErrLogFile)

	return nil
}
