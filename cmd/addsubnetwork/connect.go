package main

import (
	"fmt"
	"github.com/daglabs/btcd/rpcclient"
	"io/ioutil"
	"log"
)

func connect(cfg *config) (*rpcclient.Client, error) {
	var cert []byte
	if !cfg.DisableTLS {
		var err error
		cert, err = ioutil.ReadFile(cfg.RPCCert)
		if err != nil {
			return nil, fmt.Errorf("error reading certificates file: %s", err)
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
		return nil, fmt.Errorf("error connecting to address %s: %s", cfg.RPCServer, err)
	}

	log.Printf("Connected to server %s", cfg.RPCServer)

	return client, nil
}
