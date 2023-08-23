package main

import (
	"github.com/c4ei/kaspad/infrastructure/config"
	"github.com/jessevdk/go-flags"
)

type configFlags struct {
	config.NetworkFlags
}

func parseConfig() (*configFlags, error) {
	cfg := &configFlags{}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	err = cfg.ResolveNetwork(parser)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
