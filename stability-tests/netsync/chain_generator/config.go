package main

import (
	"path/filepath"

	"github.com/kaspanet/automation/stability-tests/common"

	"github.com/jessevdk/go-flags"
)

const (
	defaultLogFilename    = "netsync.log"
	defaultErrLogFilename = "netsync_err.log"
)

var (
	// Default configuration options
	defaultLogFile    = filepath.Join(common.DefaultHomeDir, defaultLogFilename)
	defaultErrLogFile = filepath.Join(common.DefaultHomeDir, defaultErrLogFilename)
)

type configFlags struct {
	LogLevel       string `short:"d" long:"loglevel" description:"Set log level {trace, debug, info, warn, error, critical}"`
	NumberOfBlocks uint64 `short:"n" long:"numblocks" description:"Number of blocks to mine" required:"true"`
	TargetFile     string `short:"f" long:"targetfile" description:"The target file for the JSON" required:"true"`
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

	initLog(defaultLogFile, defaultErrLogFile)

	return nil
}
