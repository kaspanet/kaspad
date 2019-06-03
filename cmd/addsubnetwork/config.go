package main

import (
	"errors"
	"fmt"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/jessevdk/go-flags"
)

type config struct {
	PrivateKey    string `short:"k" long:"private-key" description:"Private key" required:"true"`
	RPCUser       string `short:"u" long:"rpcuser" description:"RPC username" required:"true"`
	RPCPassword   string `short:"P" long:"rpcpass" default-mask:"-" description:"RPC password" required:"true"`
	RPCServer     string `short:"s" long:"rpcserver" description:"RPC server to connect to" required:"true"`
	RPCCert       string `short:"c" long:"rpccert" description:"RPC server certificate chain for validation"`
	DisableTLS    bool   `long:"notls" description:"Disable TLS"`
	TestNet       bool   `long:"testnet" description:"Connect to testnet"`
	SimNet        bool   `long:"simnet" description:"Connect to the simulation test network"`
	DevNet        bool   `long:"devnet" description:"Connect to the development test network"`
	GasLimit      uint64 `long:"gaslimit" description:"The gas limit of the new subnetwork"`
	RegistryTxFee uint64 `long:"regtxfee" description:"The fee for the subnetwork registry transaction"`
}

const (
	defaultSubnetworkGasLimit = 1000
	defaultRegistryTxFee      = 3000
)

var (
	activeNetParams dagconfig.Params
)

func parseConfig() (*config, error) {
	cfg := &config{}
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

	activeNetParams = dagconfig.MainNetParams
	switch {
	case cfg.TestNet:
		activeNetParams = dagconfig.TestNet3Params
	case cfg.SimNet:
		activeNetParams = dagconfig.SimNetParams
	case cfg.DevNet:
		activeNetParams = dagconfig.DevNetParams
	}

	if cfg.GasLimit < 0 {
		return nil, fmt.Errorf("gaslimit may not be smaller than 0")
	}
	if cfg.GasLimit == 0 {
		cfg.GasLimit = defaultSubnetworkGasLimit
	}

	if cfg.RegistryTxFee < 0 {
		return nil, fmt.Errorf("regtxfee may not be smaller than 0")
	}
	if cfg.RegistryTxFee == 0 {
		cfg.RegistryTxFee = defaultRegistryTxFee
	}

	return cfg, nil
}
