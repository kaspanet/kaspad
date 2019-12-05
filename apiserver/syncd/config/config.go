package config

import (
	"github.com/daglabs/btcd/apiserver/config"
	"github.com/daglabs/btcd/apiserver/logger"
	"github.com/daglabs/btcd/util"
	"github.com/jessevdk/go-flags"
	"path/filepath"
)

const (
	defaultLogFilename    = "syncd.log"
	defaultErrLogFilename = "syncd_err.log"
)

var (
	// Default configuration options
	defaultLogDir = util.AppDataDir("syncd", false)
	activeConfig  *Config
)

// ActiveConfig returns the active configuration struct
func ActiveConfig() *Config {
	return activeConfig
}

// Config defines the configuration options for the sync daemon.
type Config struct {
	LogDir     string `long:"logdir" description:"Directory to log output."`
	DebugLevel string `short:"d" long:"debuglevel" description:"Set log level {trace, debug, info, warn, error, critical}"`
	Migrate    bool   `long:"migrate" description:"Migrate the database to the latest version. The daemon will not start when using this flag."`
	config.ApiServerFlags
}

// Parse parses the CLI arguments and returns a config struct.
func Parse() error {
	activeConfig = &Config{
		LogDir: defaultLogDir,
	}
	parser := flags.NewParser(activeConfig, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()
	if err != nil {
		return err
	}

	err = activeConfig.ResolveApiServerFlags(parser)
	if err != nil {
		return err
	}

	logFile := filepath.Join(activeConfig.LogDir, defaultLogFilename)
	errLogFile := filepath.Join(activeConfig.LogDir, defaultErrLogFilename)
	logger.InitLog(logFile, errLogFile)

	if activeConfig.DebugLevel != "" {
		err := logger.SetLogLevels(activeConfig.DebugLevel)
		if err != nil {
			return err
		}
	}

	return nil
}
