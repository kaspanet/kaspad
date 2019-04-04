package main

import (
	"errors"

	"github.com/jessevdk/go-flags"
)

type config struct {
	AddressListPath string `long:"addresslist" description:"Path to a list of nodes' JSON-RPC endpoints" required:"true"`
	CertificatePath string `long:"cert" description:"Path to certificate accepted by JSON-RPC endpoint"`
	DisableTLS      bool   `long:"notls" description:"Disable TLS"`
}

func parseConfig() (*config, error) {
	cfg := &config{}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()

	if err != nil {
		return nil, err
	}

	if cfg.CertificatePath == "" && !cfg.DisableTLS {
		return nil, errors.New("--notls has to be disabled if --cert is used")
	}

	if cfg.CertificatePath != "" && cfg.DisableTLS {
		return nil, errors.New("--cert should be omitted if --notls is used")
	}

	return cfg, nil
}
