package config

import (
	"github.com/jessevdk/go-flags"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/kasparov/logger"
	"github.com/kaspanet/kaspad/util"
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
	KasparovdURL string  `long:"kasparovd-url" description:"The API server url to connect to"`
	PrivateKey   string  `long:"private-key" description:"Faucet Private key"`
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

	if !cfg.Migrate {
		if cfg.KasparovdURL == "" {
			return errors.New("api-server-url argument is required when --migrate flag is not raised")
		}
		if cfg.PrivateKey == "" {
			return errors.New("private-key argument is required when --migrate flag is not raised")
		}
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
