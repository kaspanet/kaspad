package main

import (
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/pkg/errors"

	"github.com/jessevdk/go-flags"
)

var (
	defaultRPCServer = "localhost"
)

type configFlags struct {
	RPCUser     string `short:"u" long:"rpcuser" description:"RPC username"`
	RPCPassword string `short:"P" long:"rpcpass" default-mask:"-" description:"RPC password"`
	RPCServer   string `short:"s" long:"rpcserver" description:"RPC server to connect to"`
	RPCCert     string `short:"c" long:"rpccert" description:"RPC server certificate chain for validation"`
	DisableTLS  bool   `long:"notls" description:"Disable TLS"`
	config.NetworkFlags
}

func parseConfig() (*configFlags, error) {
	cfg := &configFlags{
		RPCServer: defaultRPCServer,
	}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	err = cfg.ResolveNetwork(parser)
	if err != nil {
		return nil, err
	}

	if cfg.RPCUser == "" {
		return nil, errors.New("--rpcuser is required")
	}
	if cfg.RPCPassword == "" {
		return nil, errors.New("--rpcpass is required")
	}

	if cfg.RPCCert == "" && !cfg.DisableTLS {
		return nil, errors.New("either --notls or --rpccert must be specified")
	}
	if cfg.RPCCert != "" && cfg.DisableTLS {
		return nil, errors.New("--rpccert should be omitted if --notls is used")
	}

	return cfg, nil
}
