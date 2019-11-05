package main

import (
	"github.com/daglabs/btcd/cmd/config"
	"github.com/jessevdk/go-flags"
)

type configFlags struct {
	Transaction string `long:"transaction" short:"t" description:"Unsigned transaction in HEX format" required:"true"`
	PrivateKey  string `long:"private-key" short:"p" description:"Private key" required:"true"`
	config.NetworkFlags
}

func parseCommandLine() (*configFlags, error) {
	cfg := &configFlags{}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	err = config.ParseNetConfig(cfg.NetworkFlags, parser)
	return cfg, err
}
