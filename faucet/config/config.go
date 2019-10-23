package config

import (
	"github.com/daglabs/btcd/apiserver/logger"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/util"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"path/filepath"
)

const (
	defaultLogFilename    = "faucet.log"
	defaultErrLogFilename = "faucet_err.log"
)

var (
	// Default configuration options
	defaultLogDir     = util.AppDataDir("faucet", false)
	defaultDBAddress  = "localhost:3306"
	defaultHTTPListen = "0.0.0.0:8081"

	// activeNetParams are the currently active net params
	activeNetParams *dagconfig.Params
)

// Config defines the configuration options for the API server.
type Config struct {
	LogDir       string  `long:"logdir" description:"Directory to log output."`
	HTTPListen   string  `long:"listen" description:"HTTP address to listen on (default: 0.0.0.0:8081)"`
	APIServerURL string  `long:"api-server-url" description:"The API server url to connect to" required:"true"`
	PrivateKey   string  `long:"private-key" description:"Faucet Private key" required:"true"`
	DBAddress    string  `long:"dbaddress" description:"Database address"`
	DBUser       string  `long:"dbuser" description:"Database user" required:"true"`
	DBPassword   string  `long:"dbpass" description:"Database password" required:"true"`
	DBName       string  `long:"dbname" description:"Database name" required:"true"`
	Migrate      bool    `long:"migrate" description:"Migrate the database to the latest version. The server will not start when using this flag."`
	FeeRate      float64 `long:"fee-rate" description:"Coins per gram fee rate"`
	TestNet      bool    `long:"testnet" description:"Connect to testnet"`
	SimNet       bool    `long:"simnet" description:"Connect to the simulation test network"`
	DevNet       bool    `long:"devnet" description:"Connect to the development test network"`
}

var cfg *Config

// Parse parses the CLI arguments and returns a config struct.
func Parse() error {
	cfg = &Config{
		LogDir:     defaultLogDir,
		DBAddress:  defaultDBAddress,
		HTTPListen: defaultHTTPListen,
	}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()
	if err != nil {
		return err
	}

	err = resolveNetwork(cfg)
	if err != nil {
		return err
	}

	logFile := filepath.Join(cfg.LogDir, defaultLogFilename)
	errLogFile := filepath.Join(cfg.LogDir, defaultErrLogFilename)
	logger.InitLog(logFile, errLogFile)

	return nil
}

func resolveNetwork(cfg *Config) error {
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
		return errors.New("multiple net params (testnet, simnet, devnet, etc.) can't be used " +
			"together -- choose one of them")
	}

	activeNetParams = &dagconfig.MainNetParams
	switch {
	case cfg.TestNet:
		activeNetParams = &dagconfig.TestNetParams
	case cfg.SimNet:
		activeNetParams = &dagconfig.SimNetParams
	case cfg.DevNet:
		activeNetParams = &dagconfig.DevNetParams
	}

	return nil
}

// MainConfig is a getter to the main config
func MainConfig() (*Config, error) {
	if cfg == nil {
		return nil, errors.New("No configuration was set for the faucet")
	}
	return cfg, nil
}

// ActiveNetParams returns the currently active net params
func ActiveNetParams() *dagconfig.Params {
	return activeNetParams
}
