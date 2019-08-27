package main

import (
	"fmt"
	"github.com/daglabs/btcd/rpcclient"
	"io/ioutil"
)

func connectToServer(cfg *config) (*apiServerClient, error) {
	var cert []byte
	if !cfg.DisableTLS {
		var err error
		cert, err = ioutil.ReadFile(cfg.RPCCert)
		if err != nil {
			return nil, fmt.Errorf("Error reading certificates file: %s", err)
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

	client, err := newAPIServerClient(connCfg)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to address %s: %s", cfg.RPCServer, err)
	}

	log.Infof("Connected to server %s", cfg.RPCServer)

	return client, nil
}
