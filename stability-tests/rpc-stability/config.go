package main

import (
	"github.com/kaspanet/automation/stability-tests/common/rpc"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"path/filepath"

	"github.com/kaspanet/automation/stability-tests/common"
	"github.com/kaspanet/kaspad/infrastructure/config"

	"github.com/jessevdk/go-flags"
)

const (
	defaultLogFilename    = "json_stability.log"
	defaultErrLogFilename = "json_stability_err.log"
)

var (
	// Default configuration options
	defaultLogFile    = filepath.Join(common.DefaultHomeDir, defaultLogFilename)
	defaultErrLogFile = filepath.Join(common.DefaultHomeDir, defaultErrLogFilename)
)

type configFlags struct {
	rpc.RPCConfig
	config.NetworkFlags
	CommandsFilePath string `long:"commands" short:"p" description:"Path to commands file"`
	Profile          string `long:"profile" description:"Enable HTTP profiling on given port -- NOTE port must be between 1024 and 65536"`
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

	err = rpc.ValidateRPCConfig(&cfg.RPCConfig)
	if err != nil {
		return err
	}
	log.SetLevel(logger.LevelInfo)
	common.InitBackend(backendLog, defaultLogFile, defaultErrLogFile)

	return nil
}
