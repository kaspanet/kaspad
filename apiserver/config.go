package main

import (
	"errors"
	"github.com/daglabs/btcd/util"
	"path/filepath"

	"github.com/jessevdk/go-flags"
)

const (
	defaultLogFilename    = "apiserver.log"
	defaultErrLogFilename = "apiserver_err.log"
)

var (
	// Default configuration options
	defaultHomeDir    = util.AppDataDir("apiserver", false)
	defaultLogFile    = filepath.Join(defaultHomeDir, defaultLogFilename)
	defaultErrLogFile = filepath.Join(defaultHomeDir, defaultErrLogFilename)
)

type config struct {
	Address         string `long:"address" description:"An address to a JSON-RPC endpoints" required:"true"`
	CertificatePath string `long:"cert" description:"Path to certificate accepted by JSON-RPC endpoint"`
	DisableTLS      bool   `long:"notls" description:"Disable TLS"`
}

func parseConfig() (*config, error) {
	cfg := &config{}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()

	if err != nil {
		return nil, err
	}

	if cfg.CertificatePath == "" && !cfg.DisableTLS {
		return nil, errors.New("--notls has to be disabled if --cert is used")
	}

	if cfg.CertificatePath != "" && cfg.DisableTLS {
		return nil, errors.New("--cert should be omitted if --notls is used")
	}

	initLog(defaultLogFile, defaultErrLogFile)

	return cfg, nil
}
