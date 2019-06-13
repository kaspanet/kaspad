package main

import (
	"fmt"
	"github.com/daglabs/btcd/rpcclient"
	"io/ioutil"
)

func connectToServer(cfg *config) (*txgenClient, error) {

	var cert []byte
	if !cfg.DisableTLS {
		var err error
		cert, err = ioutil.ReadFile(cfg.CertificatePath)
		if err != nil {
			return nil, fmt.Errorf("Error reading certificates file: %s", err)
		}
	}

	connCfg := &rpcclient.ConnConfig{
		Host:       cfg.Address,
		Endpoint:   "ws",
		User:       "user",
		Pass:       "pass",
		DisableTLS: cfg.DisableTLS,
	}

	if !cfg.DisableTLS {
		connCfg.Certificates = cert
	}

	client, err := newTxgenClient(connCfg)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to address %s: %s", cfg.Address, err)
	}

	log.Infof("Connected to server %s", cfg.Address)

	return client, nil
}
