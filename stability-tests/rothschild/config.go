package main

import (
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
	"github.com/kaspanet/automation/stability-tests/common"
	"github.com/kaspanet/kaspad/infrastructure/config"
)

const (
	defaultLogFilename    = "rothschild.log"
	defaultErrLogFilename = "rothschild_err.log"
)

var (
	// Default configuration options
	defaultLogFile    = filepath.Join(common.DefaultHomeDir, defaultLogFilename)
	defaultErrLogFile = filepath.Join(common.DefaultHomeDir, defaultErrLogFilename)
)

type configFlags struct {
	Profile             string `long:"profile" description:"Enable HTTP profiling on given port -- NOTE port must be between 1024 and 65536"`
	RPCServer           string `long:"rpcserver" short:"s" description:"RPC server to connect to"`
	AddressesFilePath   string `long:"addresses-file" short:"a" description:"path of file containing our and everybody else's addresses'"`
	TransactionInterval uint   `long:"transaction-interval" short:"i" description:"Time between transactions (in milliseconds; default:1000)"`
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
		if err, ok := err.(*flags.Error); ok && err.Type == flags.ErrHelp {
			os.Exit(0)
		}
		return err
	}

	err = cfg.ResolveNetwork(parser)
	if err != nil {
		return err
	}

	if cfg.TransactionInterval == 0 {
		cfg.TransactionInterval = 1000
	}

	log.SetLevel(logger.LevelInfo)
	common.InitBackend(backendLog, defaultLogFile, defaultErrLogFile)

	return nil
}
