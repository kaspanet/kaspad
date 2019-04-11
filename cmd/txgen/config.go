package main

import (
	"errors"

	"github.com/jessevdk/go-flags"
)

type config struct {
	GenerateAddress bool   `long:"genaddress" description:"Generate new dagcoin address and exit"`
	AddressListPath string `long:"addresslist" description:"Path to a list of nodes' JSON-RPC endpoints"`
	PrivateKey      string `long:"private-key" description:"Private key"`
	CertificatePath string `long:"cert" description:"Path to certificate accepted by JSON-RPC endpoint"`
	DisableTLS      bool   `long:"notls" description:"Disaable TLS"`
}

func parseConfig() (*config, error) {
	cfg := &config{}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()

	if err != nil {
		return nil, err
	}

	if cfg.GenerateAddress {
		return cfg, nil
	}

	if cfg.AddressListPath == "" {
		return nil, errors.New("--addresslist is missed")
	}

	if cfg.PrivateKey == "" {
		return nil, errors.New("--private-key is missed")
	}

	if cfg.CertificatePath == "" && !cfg.DisableTLS {
		return nil, errors.New("--notls has to be disabled if --cert is used")
	}

	if cfg.CertificatePath != "" && cfg.DisableTLS {
		return nil, errors.New("--cert should be omitted if --notls is used")
	}

	if cfg.PrivateKey == "" {
		return nil, errors.New("Private key is missed")
	}

	return cfg, nil
}
