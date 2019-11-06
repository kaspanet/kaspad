package main

import (
	"github.com/daglabs/btcd/config"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
)

type configFlags struct {
	PrivateKey    string `short:"k" long:"private-key" description:"Private key" required:"true"`
	RPCUser       string `short:"u" long:"rpcuser" description:"RPC username" required:"true"`
	RPCPassword   string `short:"P" long:"rpcpass" default-mask:"-" description:"RPC password" required:"true"`
	RPCServer     string `short:"s" long:"rpcserver" description:"RPC server to connect to" required:"true"`
	RPCCert       string `short:"c" long:"rpccert" description:"RPC server certificate chain for validation"`
	DisableTLS    bool   `long:"notls" description:"Disable TLS"`
	GasLimit      uint64 `long:"gaslimit" description:"The gas limit of the new subnetwork"`
	RegistryTxFee uint64 `long:"regtxfee" description:"The fee for the subnetwork registry transaction"`
	config.NetworkFlags
}

const (
	defaultSubnetworkGasLimit = 1000
	defaultRegistryTxFee      = 3000
)

func parseConfig() (*configFlags, error) {
	cfg := &configFlags{}
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

	err = cfg.ResolveNetwork(parser)
	if err != nil {
		return nil, err
	}

	if cfg.GasLimit < 0 {
		return nil, errors.Errorf("gaslimit may not be smaller than 0")
	}
	if cfg.GasLimit == 0 {
		cfg.GasLimit = defaultSubnetworkGasLimit
	}

	if cfg.RegistryTxFee < 0 {
		return nil, errors.Errorf("regtxfee may not be smaller than 0")
	}
	if cfg.RegistryTxFee == 0 {
		cfg.RegistryTxFee = defaultRegistryTxFee
	}

	return cfg, nil
}
