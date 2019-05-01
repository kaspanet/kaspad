package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/daglabs/btcd/rpcclient"
)

func connectToServers(cfg *config, addressList []string) ([]*simulatorClient, error) {
	clients := make([]*simulatorClient, len(addressList))

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
			Host:           address,
			Endpoint:       "ws",
			User:           "user",
			Pass:           "pass",
			DisableTLS:     cfg.DisableTLS,
			RequestTimeout: time.Second / 2,
		}

		if !cfg.DisableTLS {
			connCfg.Certificates = cert
		}

		var err error
		clients[i], err = newSimulatorClient(address, connCfg)
		if err != nil {
			return nil, err
		}

		log.Infof("Connected to server %s", address)
	}

	return clients, nil
}
