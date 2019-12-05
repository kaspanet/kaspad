package config

import (
	"github.com/daglabs/btcd/config"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
)

var (
	// Default configuration options
	defaultDBAddress = "localhost:3306"
)

// ApiServerFlags holds configuration common to both the server and the daemon.
type ApiServerFlags struct {
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

// ResolveApiServerFlags parses command line arguments and sets ApiServerFlags accordingly.
func (apiServerFlags *ApiServerFlags) ResolveApiServerFlags(parser *flags.Parser) error {
	if apiServerFlags.DBAddress == "" {
		apiServerFlags.DBAddress = defaultDBAddress
	}
	if apiServerFlags.RPCUser == "" {
		return errors.New("--rpcuser is required")
	}
	if apiServerFlags.RPCPassword == "" {
		return errors.New("--rpcpass is required")
	}
	if apiServerFlags.RPCServer == "" {
		return errors.New("--rpcserver is required")
	}

	if apiServerFlags.RPCCert == "" && !apiServerFlags.DisableTLS {
		return errors.New("--notls has to be disabled if --cert is used")
	}

	if apiServerFlags.RPCCert != "" && apiServerFlags.DisableTLS {
		return errors.New("--cert should be omitted if --notls is used")
	}
	return apiServerFlags.ResolveNetwork(parser)
}
