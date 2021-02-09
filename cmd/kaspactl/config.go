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
	RPCServer            string `short:"s" long:"rpcserver" description:"RPC server to connect to"`
	Timeout              uint64 `short:"t" long:"timeout" description:"Timeout for the request (in seconds)"`
	RequestJSON          string `short:"j" long:"json" description:"The request in JSON format"`
	ListCommands         bool   `short:"l" long:"list-commands" description:"List all commands and exit"`
	CommandAndParameters []string
	config.NetworkFlags
}

func parseConfig() (*configFlags, error) {
	cfg := &configFlags{
		RPCServer: defaultRPCServer,
		Timeout:   defaultTimeout,
	}
	parser := flags.NewParser(cfg, flags.HelpFlag)
	parser.Usage = "kaspactl [OPTIONS] [COMMAND] [COMMAND PARAMETERS].\n\nCommand can be supplied only if --json is not used." +
		"\n\nUse `kaspactl --list-commands` to get a list of all commands and their parameters"
	remainingArgs, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	if cfg.ListCommands {
		return cfg, nil
	}

	err = cfg.ResolveNetwork(parser)
	if err != nil {
		return nil, err
	}

	cfg.CommandAndParameters = remainingArgs
	if len(cfg.CommandAndParameters) == 0 && cfg.RequestJSON == "" ||
		len(cfg.CommandAndParameters) > 0 && cfg.RequestJSON != "" {

		return nil, errors.New("Exactly one of --json or a command must be specified")
	}

	return cfg, nil
}
