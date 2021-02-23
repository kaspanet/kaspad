package main

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"
	"time"
)

const minerTimeout = 10 * time.Second

type minerClient struct {
	isReconnecting uint32
	clientLock     sync.RWMutex
	rpcClient      *rpcclient.RPCClient

	cfg                        *configFlags
	blockAddedNotificationChan chan struct{}
}

func (mc *minerClient) safeRPCClient() *rpcclient.RPCClient {
	mc.clientLock.RLock()
	defer mc.clientLock.RUnlock()
	return mc.rpcClient
}

func (mc *minerClient) reconnect() {
	swapped := atomic.CompareAndSwapUint32(&mc.isReconnecting, 0, 1)
	if !swapped {
		return
	}

	defer atomic.StoreUint32(&mc.isReconnecting, 0)

	mc.clientLock.Lock()
	defer mc.clientLock.Unlock()

	retryDuration := time.Second
	log.Infof("Reconnecting RPC connection")
	for {
		err := mc.connect()
		if err == nil {
			return
		}

		const maxRetryDuration = time.Minute
		if retryDuration < time.Minute {
			retryDuration *= 2
		} else {
			retryDuration = maxRetryDuration
		}

		log.Errorf("Got error %s while reconnecting. Trying again in %s", err, retryDuration)
		time.Sleep(retryDuration)
	}
}

func (mc *minerClient) connect() error {
	rpcAddress, err := mc.cfg.NetParams().NormalizeRPCServerAddress(mc.cfg.RPCServer)
	if err != nil {
		return err
	}
	mc.rpcClient, err = rpcclient.NewRPCClient(rpcAddress)
	if err != nil {
		return err
	}
	mc.rpcClient.SetTimeout(minerTimeout)
	mc.rpcClient.SetLogger(backendLog, logger.LevelTrace)

	err = mc.rpcClient.RegisterForBlockAddedNotifications(func(_ *appmessage.BlockAddedNotificationMessage) {
		select {
		case mc.blockAddedNotificationChan <- struct{}{}:
		default:
		}
	})
	if err != nil {
		return errors.Wrapf(err, "error requesting block-added notifications")
	}

	log.Infof("Connected to %s", rpcAddress)

	return nil
}

func newMinerClient(cfg *configFlags) (*minerClient, error) {
	minerClient := &minerClient{
		cfg:                        cfg,
		blockAddedNotificationChan: make(chan struct{}),
	}

	err := minerClient.connect()
	if err != nil {
		return nil, err
	}

	return minerClient, nil
}
