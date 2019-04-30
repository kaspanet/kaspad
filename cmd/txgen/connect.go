package main

import (
	"fmt"
	"github.com/daglabs/btcd/rpcclient"
	"io/ioutil"
)

func connectToServers(cfg *config, addressList []string) ([]*rpcclient.Client, error) {
	clients := make([]*rpcclient.Client, len(addressList))

	var cert []byte
	if !cfg.DisableTLS {
		var err error
		cert, err = ioutil.ReadFile(cfg.CertificatePath)
		if err != nil {
			return nil, fmt.Errorf("Error reading certificates file: %s", err)
		}
	}

	for i, address := range addressList {
		connCfg := &rpcclient.ConnConfig{
			Host:       address,
			Endpoint:   "ws",
			User:       "user",
			Pass:       "pass",
			DisableTLS: cfg.DisableTLS,
		}

		if !cfg.DisableTLS {
			connCfg.Certificates = cert
		}

		client, err := rpcclient.New(connCfg, nil)
		if err != nil {
			return nil, fmt.Errorf("Error connecting to address %s: %s", address, err)
		}

		clients[i] = client

		logger.Infof("Connected to server %s", address)
	}

	return clients, nil
}
