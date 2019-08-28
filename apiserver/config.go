package main

import (
	"errors"
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
	defaultLogDir    = util.AppDataDir("apiserver", false)
	defaultDBAddress = "localhost:3306"
)

type config struct {
	LogDir      string `long:"logdir" description:"Directory to log output."`
	RPCUser     string `short:"u" long:"rpcuser" description:"RPC username" required:"true"`
	RPCPassword string `short:"P" long:"rpcpass" default-mask:"-" description:"RPC password" required:"true"`
	RPCServer   string `short:"s" long:"rpcserver" description:"RPC server to connect to" required:"true"`
	RPCCert     string `short:"c" long:"rpccert" description:"RPC server certificate chain for validation"`
	DisableTLS  bool   `long:"notls" description:"Disable TLS"`
	DBAddress   string `long:"dbaddress" description:"Database host and port"`
	DBUser      string `long:"dbuser" description:"Database user" required:"true"`
	DBPassword  string `long:"dbpass" description:"Database password" required:"true"`
	DBName      string `long:"dbname" description:"Database name" required:"true"`
}

func parseConfig() (*config, error) {
	cfg := &config{
		LogDir:    defaultLogDir,
		DBAddress: defaultDBAddress,
	}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()

	if err != nil {
		return nil, err
	}

	if cfg.RPCCert == "" && !cfg.DisableTLS {
		return nil, errors.New("--notls has to be disabled if --cert is used")
	}

	if cfg.RPCCert != "" && cfg.DisableTLS {
		return nil, errors.New("--cert should be omitted if --notls is used")
	}

	logFile := filepath.Join(cfg.LogDir, defaultLogFilename)
	errLogFile := filepath.Join(cfg.LogDir, defaultErrLogFilename)
	initLog(logFile, errLogFile)

	return cfg, nil
}
