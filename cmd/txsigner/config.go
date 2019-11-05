package main

import (
	"github.com/daglabs/btcd/cmd/cmdconfig"
	"github.com/jessevdk/go-flags"
)

type config struct {
	Transaction string `long:"transaction" short:"t" description:"Unsigned transaction in HEX format" required:"true"`
	PrivateKey  string `long:"private-key" short:"p" description:"Private key" required:"true"`
	cmdconfig.NetConfig
}

func parseCommandLine() (*config, error) {
	cfg := &config{}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	err = cmdconfig.ParseNetConfig(cfg.NetConfig, parser)
	return cfg, err
}
