package daa

import (
	"path/filepath"

	"github.com/jessevdk/go-flags"
	"github.com/kaspanet/kaspad/stability-tests/common"
)

const (
	defaultLogFilename    = "daa.log"
	defaultErrLogFilename = "daa_err.log"
)

var (
	// Default configuration options
	defaultLogFile    = filepath.Join(common.DefaultAppDir, defaultLogFilename)
	defaultErrLogFile = filepath.Join(common.DefaultAppDir, defaultErrLogFilename)
)

type configFlags struct {
	LogLevel string `long:"loglevel" description:"Set log level {trace, debug, info, warn, error, critical}"`
	Profile  string `long:"profile" description:"Enable HTTP profiling on given port -- NOTE port must be between 1024 and 65536"`
}

var cfg *configFlags

func activeConfig() *configFlags {
	return cfg
}

func parseConfig() error {
	cfg = &configFlags{}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag|flags.IgnoreUnknown)
	_, err := parser.Parse()
	if err != nil {
		return err
	}

	initLog(defaultLogFile, defaultErrLogFile)

	return nil
}
