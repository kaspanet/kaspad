package main

import (
	"path/filepath"

	"github.com/zoomy-network/zoomyd/infrastructure/config"
	"github.com/zoomy-network/zoomyd/infrastructure/logger"
	"github.com/zoomy-network/zoomyd/stability-tests/common"
	"github.com/zoomy-network/zoomyd/stability-tests/common/rpc"

	"github.com/jessevdk/go-flags"
)

const (
	defaultLogFilename    = "json_stability.log"
	defaultErrLogFilename = "json_stability_err.log"
)

var (
	// Default configuration options
	defaultLogFile    = filepath.Join(common.DefaultAppDir, defaultLogFilename)
	defaultErrLogFile = filepath.Join(common.DefaultAppDir, defaultErrLogFilename)
)

type configFlags struct {
	rpc.Config
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

	err = rpc.ValidateRPCConfig(&cfg.Config)
	if err != nil {
		return err
	}
	log.SetLevel(logger.LevelInfo)
	common.InitBackend(backendLog, defaultLogFile, defaultErrLogFile)

	return nil
}
