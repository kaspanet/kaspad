package config

import (
	"github.com/daglabs/btcd/kasparov/config"
	"github.com/daglabs/btcd/kasparov/logger"
	"github.com/daglabs/btcd/util"
	"github.com/jessevdk/go-flags"
	"path/filepath"
)

const (
	defaultLogFilename    = "apiserver.log"
	defaultErrLogFilename = "apiserver_err.log"
)

var (
	// Default configuration options
	defaultLogDir     = util.AppDataDir("apiserver", false)
	defaultHTTPListen = "0.0.0.0:8080"
	activeConfig      *Config
)

// ActiveConfig returns the active configuration struct
func ActiveConfig() *Config {
	return activeConfig
}

// Config defines the configuration options for the API server.
type Config struct {
	LogDir     string `long:"logdir" description:"Directory to log output."`
	DebugLevel string `short:"d" long:"debuglevel" description:"Set log level {trace, debug, info, warn, error, critical}"`
	HTTPListen string `long:"listen" description:"HTTP address to listen on (default: 0.0.0.0:8080)"`
	config.KasparovFlags
}

// Parse parses the CLI arguments and returns a config struct.
func Parse() error {
	activeConfig = &Config{
		LogDir:     defaultLogDir,
		HTTPListen: defaultHTTPListen,
	}
	parser := flags.NewParser(activeConfig, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()
	if err != nil {
		return err
	}

	err = activeConfig.ResolveKasparovFlags(parser)
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
