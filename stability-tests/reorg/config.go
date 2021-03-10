package main

import (
	"os"
	"path/filepath"

	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/stability-tests/common"

	"github.com/jessevdk/go-flags"
)

const (
	defaultLogFilename    = "reorg.log"
	defaultErrLogFilename = "reorg_err.log"
)

var (
	// Default configuration options
	defaultLogFile    = filepath.Join(common.DefaultHomeDir, defaultLogFilename)
	defaultErrLogFile = filepath.Join(common.DefaultHomeDir, defaultErrLogFilename)
)

type configFlags struct {
	Profile string `long:"profile" description:"Enable HTTP profiling on given port -- NOTE port must be between 1024 and 65536"`
	DAGFile string `long:"dag-file" description:"Path to DAG JSON file"`
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
		if err, ok := err.(*flags.Error); ok && err.Type == flags.ErrHelp {
			os.Exit(0)
		}
		return err
	}
	log.SetLevel(logger.LevelInfo)
	common.InitBackend(backendLog, defaultLogFile, defaultErrLogFile)

	return nil
}
