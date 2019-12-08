package config

import (
	"github.com/daglabs/kaspad/kasparov/config"
	"github.com/daglabs/kaspad/util"
	"github.com/jessevdk/go-flags"
)

const (
	logFilename    = "apiserver.log"
	errLogFilename = "apiserver_err.log"
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
	HTTPListen string `long:"listen" description:"HTTP address to listen on (default: 0.0.0.0:8080)"`
	config.KasparovFlags
}

// Parse parses the CLI arguments and returns a config struct.
func Parse() error {
	activeConfig = &Config{
		HTTPListen: defaultHTTPListen,
	}
	parser := flags.NewParser(activeConfig, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()
	if err != nil {
		return err
	}

	err = activeConfig.ResolveKasparovFlags(parser, defaultLogDir, logFilename, errLogFilename)
	if err != nil {
		return err
	}

	return nil
}
