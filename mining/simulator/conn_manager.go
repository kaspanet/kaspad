package main

import (
	"io/ioutil"
	"time"

	"github.com/daglabs/btcd/rpcclient"
	"github.com/pkg/errors"
)

type connectionManager struct {
	addressList    []string
	cert           []byte
	clients        []*simulatorClient
	cfg            *config
	disconnectChan chan struct{}
}

func newConnectionManager(cfg *config) (*connectionManager, error) {
	connManager := &connectionManager{}
	var err error

	connManager.addressList, err = getAddressList(cfg)
	if err != nil {
		return nil, err
	}

	connManager.cert, err = readCert(cfg)
	if err != nil {
		return nil, err
	}

	connManager.clients, err = connectToServers(connManager.addressList, connManager.cert)
	if err != nil {
		return nil, err
	}

	if cfg.AutoScalingGroup != "" {
		connManager.disconnectChan = make(chan struct{})
		spawn(func() { connManager.refreshAddressesLoop() })
	}

	return connManager, nil
}

func connectToServer(address string, cert []byte) (*simulatorClient, error) {
	connCfg := &rpcclient.ConnConfig{
		Host:           address,
		Endpoint:       "ws",
		User:           "user",
		Pass:           "pass",
		DisableTLS:     cert == nil,
		RequestTimeout: time.Second * 10,
		Certificates:   cert,
	}

	client, err := newSimulatorClient(address, connCfg)
	if err != nil {
		return nil, err
	}

	log.Infof("Connected to server %s", address)

	return client, nil
}

func connectToServers(addressList []string, cert []byte) ([]*simulatorClient, error) {
	clients := make([]*simulatorClient, 0, len(addressList))

	for _, address := range addressList {
		client, err := connectToServer(address, cert)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	return clients, nil
}

func readCert(cfg *config) ([]byte, error) {
	var cert []byte
	if !cfg.DisableTLS {
		var err error
		cert, err = ioutil.ReadFile(cfg.CertificatePath)
		if err != nil {
			return nil, errors.Errorf("Error reading certificates file: %s", err)
		}
	}

	return cert, nil
}

func (cm *connectionManager) close() {
	if cm.disconnectChan != nil {
		cm.disconnectChan <- struct{}{}
	}
	for _, client := range cm.clients {
		client.Disconnect()
	}
}

const refreshAddressInterval = time.Minute * 10

func (cm *connectionManager) refreshAddressesLoop() {
	for {
		select {
		case <-time.After(refreshAddressInterval):
			err := cm.refreshAddresses()
			if err != nil {
				panic(err)
			}
		case <-cm.disconnectChan:
			return
		}
	}
}

func (cm *connectionManager) refreshAddresses() error {
	newAddressList, err := getAddressList(cm.cfg)
	if err != nil {
		return err
	}

	if len(newAddressList) == len(cm.addressList) {
		return nil
	}

outerLoop:
	for _, newAddress := range newAddressList {
		for _, oldAddress := range cm.addressList {
			if newAddress == oldAddress {
				continue outerLoop
			}
		}

		client, err := connectToServer(newAddress, cm.cert)
		if err != nil {
			return err
		}
		cm.clients = append(cm.clients, client)
	}

	cm.addressList = newAddressList

	return nil
}
