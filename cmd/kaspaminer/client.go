package main

import (
	"github.com/kaspanet/kaspad/network/domainmessage"
	"github.com/kaspanet/kaspad/network/rpc/client"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"io/ioutil"
	"time"
)

type minerClient struct {
	*client.Client
	onBlockAdded chan struct{}
}

func newMinerClient(connCfg *client.ConnConfig) (*minerClient, error) {
	minerClient := &minerClient{
		onBlockAdded: make(chan struct{}, 1),
	}
	notificationHandlers := &client.NotificationHandlers{
		OnFilteredBlockAdded: func(_ uint64, header *domainmessage.BlockHeader,
			txs []*util.Tx) {
			minerClient.onBlockAdded <- struct{}{}
		},
	}
	var err error
	minerClient.Client, err = client.New(connCfg, notificationHandlers)
	if err != nil {
		return nil, errors.Errorf("Error connecting to address %s: %s", connCfg.Host, err)
	}

	if err = minerClient.NotifyBlocks(); err != nil {
		return nil, errors.Wrapf(err, "error while registering minerClient %s for block notifications", minerClient.Host())
	}
	return minerClient, nil
}

func connectToServer(cfg *configFlags) (*minerClient, error) {
	cert, err := readCert(cfg)
	if err != nil {
		return nil, err
	}

	rpcAddr, err := cfg.NetParams().NormalizeRPCServerAddress(cfg.RPCServer)
	if err != nil {
		return nil, err
	}

	connCfg := &client.ConnConfig{
		Host:           rpcAddr,
		Endpoint:       "ws",
		User:           cfg.RPCUser,
		Pass:           cfg.RPCPassword,
		DisableTLS:     cfg.DisableTLS,
		RequestTimeout: time.Second * 10,
		Certificates:   cert,
	}

	client, err := newMinerClient(connCfg)
	if err != nil {
		return nil, err
	}

	log.Infof("Connected to server %s", client.Host())

	return client, nil
}

func readCert(cfg *configFlags) ([]byte, error) {
	if cfg.DisableTLS {
		return nil, nil
	}

	cert, err := ioutil.ReadFile(cfg.RPCCert)
	if err != nil {
		return nil, errors.Errorf("Error reading certificates file: %s", err)
	}

	return cert, nil
}
