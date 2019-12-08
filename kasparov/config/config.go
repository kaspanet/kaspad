package config

import (
	"github.com/daglabs/kaspad/config"
	"github.com/daglabs/kaspad/kasparov/logger"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"path/filepath"
)

var (
	// Default configuration options
	defaultDBAddress = "localhost:3306"
)

// KasparovFlags holds configuration common to both the Kasparov server and the Kasparov daemon.
type KasparovFlags struct {
	LogDir      string `long:"logdir" description:"Directory to log output."`
	DebugLevel  string `short:"d" long:"debuglevel" description:"Set log level {trace, debug, info, warn, error, critical}"`
	DBAddress   string `long:"dbaddress" description:"Database address"`
	DBUser      string `long:"dbuser" description:"Database user" required:"true"`
	DBPassword  string `long:"dbpass" description:"Database password" required:"true"`
	DBName      string `long:"dbname" description:"Database name" required:"true"`
	RPCUser     string `short:"u" long:"rpcuser" description:"RPC username"`
	RPCPassword string `short:"P" long:"rpcpass" default-mask:"-" description:"RPC password"`
	RPCServer   string `short:"s" long:"rpcserver" description:"RPC server to connect to"`
	RPCCert     string `short:"c" long:"rpccert" description:"RPC server certificate chain for validation"`
	DisableTLS  bool   `long:"notls" description:"Disable TLS"`
	config.NetworkFlags
}

// ResolveKasparovFlags parses command line arguments and sets KasparovFlags accordingly.
func (kasparovFlags *KasparovFlags) ResolveKasparovFlags(parser *flags.Parser,
	defaultLogDir, logFilename, errLogFilename string) error {
	if kasparovFlags.LogDir == "" {
		kasparovFlags.LogDir = defaultLogDir
	}
	logFile := filepath.Join(kasparovFlags.LogDir, logFilename)
	errLogFile := filepath.Join(kasparovFlags.LogDir, errLogFilename)
	logger.InitLog(logFile, errLogFile)

	if kasparovFlags.DebugLevel != "" {
		err := logger.SetLogLevels(kasparovFlags.DebugLevel)
		if err != nil {
			return err
		}
	}

	if kasparovFlags.DBAddress == "" {
		kasparovFlags.DBAddress = defaultDBAddress
	}
	if kasparovFlags.RPCUser == "" {
		return errors.New("--rpcuser is required")
	}
	if kasparovFlags.RPCPassword == "" {
		return errors.New("--rpcpass is required")
	}
	if kasparovFlags.RPCServer == "" {
		return errors.New("--rpcserver is required")
	}

	if kasparovFlags.RPCCert == "" && !kasparovFlags.DisableTLS {
		return errors.New("--notls has to be disabled if --cert is used")
	}
	if kasparovFlags.RPCCert != "" && kasparovFlags.DisableTLS {
		return errors.New("--cert should be omitted if --notls is used")
	}
	return kasparovFlags.ResolveNetwork(parser)
}
