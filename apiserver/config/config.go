package config

import (
	"github.com/daglabs/btcd/apiserver/logger"
	"github.com/daglabs/btcd/config"
	"github.com/daglabs/btcd/util"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"path/filepath"
)

const (
	defaultLogFilename    = "apiserver.log"
	defaultErrLogFilename = "apiserver_err.log"
)

var (
	// Default configuration options
	defaultLogDir     = util.AppDataDir("apiserver", false)
	defaultDBAddress  = "localhost:3306"
	defaultHTTPListen = "0.0.0.0:8080"
	activeConfig      *Config
)

// ActiveConfig returns the active configuration struct
func ActiveConfig() *Config {
	return activeConfig
}

// Config defines the configuration options for the API server.
type Config struct {
	LogDir            string `long:"logdir" description:"Directory to log output."`
	DebugLevel        string `short:"d" long:"debuglevel" description:"Set log level {trace, debug, info, warn, error, critical}"`
	RPCUser           string `short:"u" long:"rpcuser" description:"RPC username"`
	RPCPassword       string `short:"P" long:"rpcpass" default-mask:"-" description:"RPC password"`
	RPCServer         string `short:"s" long:"rpcserver" description:"RPC server to connect to"`
	RPCCert           string `short:"c" long:"rpccert" description:"RPC server certificate chain for validation"`
	DisableTLS        bool   `long:"notls" description:"Disable TLS"`
	DBAddress         string `long:"dbaddress" description:"Database address"`
	DBUser            string `long:"dbuser" description:"Database user" required:"true"`
	DBPassword        string `long:"dbpass" description:"Database password" required:"true"`
	DBName            string `long:"dbname" description:"Database name" required:"true"`
	HTTPListen        string `long:"listen" description:"HTTP address to listen on (default: 0.0.0.0:8080)"`
	Migrate           bool   `long:"migrate" description:"Migrate the database to the latest version. The server will not start when using this flag."`
	MQTTBrokerAddress string `long:"mqttaddress" description:"MQTT broker address" required:"false"`
	MQTTUser          string `long:"mqttuser" description:"MQTT server user" required:"false"`
	MQTTPassword      string `long:"mqttpass" description:"MQTT server password" required:"false"`
	config.NetworkFlags
}

// Parse parses the CLI arguments and returns a config struct.
func Parse() error {
	activeConfig = &Config{
		LogDir:     defaultLogDir,
		DBAddress:  defaultDBAddress,
		HTTPListen: defaultHTTPListen,
	}
	parser := flags.NewParser(activeConfig, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()
	if err != nil {
		return err
	}

	if !activeConfig.Migrate {
		if activeConfig.RPCUser == "" {
			return errors.New("--rpcuser is required if --migrate flag is not used")
		}
		if activeConfig.RPCPassword == "" {
			return errors.New("--rpcpass is required if --migrate flag is not used")
		}
		if activeConfig.RPCServer == "" {
			return errors.New("--rpcserver is required if --migrate flag is not used")
		}
	}

	if activeConfig.RPCCert == "" && !activeConfig.DisableTLS {
		return errors.New("--notls has to be disabled if --cert is used")
	}

	if activeConfig.RPCCert != "" && activeConfig.DisableTLS {
		return errors.New("--cert should be omitted if --notls is used")
	}

	if (activeConfig.MQTTBrokerAddress != "" || activeConfig.MQTTUser != "" || activeConfig.MQTTPassword != "") &&
		(activeConfig.MQTTBrokerAddress == "" || activeConfig.MQTTUser == "" || activeConfig.MQTTPassword == "") {
		return errors.New("--mqttaddress, --mqttuser, and --mqttpass must be passed all together")
	}

	err = activeConfig.ResolveNetwork(parser)
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
