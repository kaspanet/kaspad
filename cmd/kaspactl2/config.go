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
	RPCServer   string `short:"s" long:"rpcserver" description:"RPC server to connect to"`
	RequestJSON string `description:"The request in JSON format"`
	config.NetworkFlags
}

func parseConfig() (*configFlags, error) {
	cfg := &configFlags{
		RPCServer: defaultRPCServer,
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
