package main

import (
	"github.com/kaspanet/kaspad/rpcclient"
	"github.com/pkg/errors"
	"io/ioutil"
)

func connect(cfg *ConfigFlags) (*rpcclient.Client, error) {
	var cert []byte
	if !cfg.DisableTLS {
		var err error
		cert, err = ioutil.ReadFile(cfg.RPCCert)
		if err != nil {
			return nil, errors.Errorf("error reading certificates file: %s", err)
		}
	}

	connCfg := &rpcclient.ConnConfig{
		Host:       cfg.RPCServer,
		Endpoint:   "ws",
		User:       cfg.RPCUser,
		Pass:       cfg.RPCPassword,
		DisableTLS: cfg.DisableTLS,
	}

	if !cfg.DisableTLS {
		connCfg.Certificates = cert
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, errors.Errorf("error connecting to address %s: %s", cfg.RPCServer, err)
	}

	return client, nil
}
