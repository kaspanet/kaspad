package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/rpcclient"
)

type simulatorClient struct {
	*rpcclient.Client
	onBlockAdded chan struct{}
}

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
		onBlockAdded := make(chan struct{}, 1)
		ntfnHandlers := &rpcclient.NotificationHandlers{
			OnBlockAdded: func(hash *daghash.Hash, height int32, t time.Time) {
				onBlockAdded <- struct{}{}
			},
		}
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

		client, err := rpcclient.New(connCfg, ntfnHandlers)
		if err != nil {
			return nil, fmt.Errorf("Error connecting to address %s: %s", address, err)
		}

		if err := client.NotifyBlocks(); err != nil {
			return nil, fmt.Errorf("Error while registering client %s for block notifications: %s", client.Host(), err)
		}

		clients[i] = &simulatorClient{
			Client:       client,
			onBlockAdded: onBlockAdded,
		}

		log.Printf("Connected to server %s", address)
	}

	return clients, nil
}
