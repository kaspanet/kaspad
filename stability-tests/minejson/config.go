package main

import (
	"path/filepath"

	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/stability-tests/common/rpc"

	"github.com/jessevdk/go-flags"
)

const (
	defaultLogFilename    = "minejson.log"
	defaultErrLogFilename = "minejson_err.log"
)

var (
	// Default configuration options
	defaultLogFile    = filepath.Join(common.DefaultAppDir, defaultLogFilename)
	defaultErrLogFile = filepath.Join(common.DefaultAppDir, defaultErrLogFilename)
)

type configFlags struct {
	rpc.Config
	DAGFile string `long:"dag-file" description:"Path to DAG JSON file"`
	Profile string `long:"profile" description:"Enable HTTP profiling on given port -- NOTE port must be between 1024 and 65536"`
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

	err = rpc.ValidateRPCConfig(&cfg.Config)
	if err != nil {
		return err
	}

	log.SetLevel(logger.LevelInfo)
	common.InitBackend(backendLog, defaultLogFile, defaultErrLogFile)
	return nil
}
