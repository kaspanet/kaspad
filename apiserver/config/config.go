package config

import (
	"errors"
	"github.com/daglabs/btcd/apiserver/logger"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/util"
	"github.com/jessevdk/go-flags"
	"path/filepath"
)

const (
	defaultLogFilename    = "apiserver.log"
	defaultErrLogFilename = "apiserver_err.log"
)

var (
	ActiveNetParams dagconfig.Params
)

var (
	// Default configuration options
	defaultLogDir     = util.AppDataDir("apiserver", false)
	defaultDBAddress  = "localhost:3306"
	defaultHTTPListen = "0.0.0.0:8080"
)

// Config defines the configuration options for the API server.
type Config struct {
	LogDir      string `long:"logdir" description:"Directory to log output."`
	RPCUser     string `short:"u" long:"rpcuser" description:"RPC username" required:"true"`
	RPCPassword string `short:"P" long:"rpcpass" default-mask:"-" description:"RPC password" required:"true"`
	RPCServer   string `short:"s" long:"rpcserver" description:"RPC server to connect to" required:"true"`
	RPCCert     string `short:"c" long:"rpccert" description:"RPC server certificate chain for validation"`
	DisableTLS  bool   `long:"notls" description:"Disable TLS"`
	DBAddress   string `long:"dbaddress" description:"Database address"`
	DBUser      string `long:"dbuser" description:"Database user" required:"true"`
	DBPassword  string `long:"dbpass" description:"Database password" required:"true"`
	DBName      string `long:"dbname" description:"Database name" required:"true"`
	HTTPListen  string `long:"listen" description:"HTTP address to listen on (default: 0.0.0.0:8080)"`
	TestNet     bool   `long:"testnet" description:"Connect to testnet"`
	SimNet      bool   `long:"simnet" description:"Connect to the simulation test network"`
	DevNet      bool   `long:"devnet" description:"Connect to the development test network"`
}

// Parse parses the CLI arguments and returns a config struct.
func Parse() (*Config, error) {
	cfg := &Config{
		LogDir:     defaultLogDir,
		DBAddress:  defaultDBAddress,
		HTTPListen: defaultHTTPListen,
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

	// Multiple networks can't be selected simultaneously.
	numNets := 0
	if cfg.TestNet {
		numNets++
	}
	if cfg.SimNet {
		numNets++
	}
	if cfg.DevNet {
		numNets++
	}
	if numNets > 1 {
		return nil, errors.New("multiple net params (testnet, simnet, devnet, etc.) can't be used " +
			"together -- choose one of them")
	}

	ActiveNetParams = dagconfig.MainNetParams
	switch {
	case cfg.TestNet:
		ActiveNetParams = dagconfig.TestNet3Params
	case cfg.SimNet:
		ActiveNetParams = dagconfig.SimNetParams
	case cfg.DevNet:
		ActiveNetParams = dagconfig.DevNetParams
	}

	logFile := filepath.Join(cfg.LogDir, defaultLogFilename)
	errLogFile := filepath.Join(cfg.LogDir, defaultErrLogFilename)
	logger.InitLog(logFile, errLogFile)

	return cfg, nil
}
