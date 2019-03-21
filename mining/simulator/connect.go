package main

import (
	"fmt"
	"io/ioutil"

	"github.com/daglabs/btcd/rpcclient"
)

const certificatePath = "rpc.cert"

func connectToServers(addressList []string) ([]*rpcclient.Client, error) {
	clients := make([]*rpcclient.Client, len(addressList))

	cert, err := ioutil.ReadFile(certificatePath)
	if err != nil {
		return nil, fmt.Errorf("Error reading certificates file: %s", err)
	}

	for i, address := range addressList {
		connCfg := &rpcclient.ConnConfig{
			Host:         address,
			Endpoint:     "ws",
			User:         "user",
			Pass:         "pass",
			Certificates: cert,
		}

		client, err := rpcclient.New(connCfg, nil)
		if err != nil {
			return nil, fmt.Errorf("Error connecting to address %s: %s", address, err)
		}

		clients[i] = client
	}

	return clients, nil
}
