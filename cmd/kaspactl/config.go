package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/pkg/errors"
)

var (
	defaultRPCServer        = "localhost"
	defaultTimeout   uint64 = 30
)

type configFlags struct {
	RPCServer   string `short:"s" long:"rpcserver" description:"RPC server to connect to"`
	Timeout     uint64 `short:"t" long:"timeout" description:"Timeout for the request (in seconds)"`
	RequestJSON string `description:"The request in JSON format"`
	config.NetworkFlags
}

func parseConfig() (*configFlags, error) {
	cfg := &configFlags{
		RPCServer: defaultRPCServer,
		Timeout:   defaultTimeout,
	}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)
	args, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	err = cfg.ResolveNetwork(parser)
	if err != nil {
		return nil, err
	}

	if len(args) != 1 {
		return nil, errors.New("the last parameter must be the request in JSON format")
	}
	cfg.RequestJSON = args[0]

	return cfg, nil
}
