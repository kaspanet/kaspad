package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/kaspanet/kaspad/config"
)

var activeConfig *ConfigFlags

// ActiveConfig returns the active configuration struct
func ActiveConfig() *ConfigFlags {
	return activeConfig
}

// ConfigFlags holds the configurations set by the command line argument
type ConfigFlags struct {
	Transaction string `long:"transaction" short:"t" description:"Unsigned transaction in HEX format" required:"true"`
	PrivateKey  string `long:"private-key" short:"p" description:"Private key" required:"true"`
	config.NetworkFlags
}

func parseCommandLine() (*ConfigFlags, error) {
	activeConfig = &ConfigFlags{}
	parser := flags.NewParser(activeConfig, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	err = activeConfig.ResolveNetwork(parser)
	if err != nil {
		return nil, err
	}

	return activeConfig, nil
}
