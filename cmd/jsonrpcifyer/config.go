package main

import (
	"errors"

	"github.com/jessevdk/go-flags"
)

type config struct {
	Host       string `long:"host" default:"localhost:18334" description:"IP:Port of the JSON-RPC endpoint"`
	ListenPort int    `long:"port" default:"8080" description:"Port to listen on"`
	RPCCert    string `long:"rpccert" description:"Path to certificate accepted by JSON-RPC endpoint"`
	RPCUser    string `long:"rpcuser" required:"true" description:"Username to connect to JSON-RPC endpoint"`
	RPCPass    string `long:"rpcpass" required:"true" description:"Password to connect to JSON-RPC endpoint"`
	DisableTLS bool   `long:"notls" description:"Disable TLS"`
}

func parseConfig() (*config, error) {
	cfg := &config{}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()

	if err != nil {
		return nil, err
	}

	if cfg.RPCCert == "" && !cfg.DisableTLS {
		return nil, errors.New("either --notls or --rpccert must be set")
	}

	return cfg, nil
}
