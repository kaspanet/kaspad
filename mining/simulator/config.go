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
		return nil, errors.New("TLS has to be disabled if no certificate is provided")
	}

	if cfg.CertificatePath != "" && cfg.DisableTLS {
		return nil, errors.New("The certificate path should be omitted if TLS is disabled")
	}

	return cfg, nil

}
