package main

import (
	"fmt"
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/rpcclient"
	"github.com/daglabs/btcd/util/subnetworkid"
	"time"
)

const (
	getSubnetworkRetryDelay = 5 * time.Second
	maxGetSubnetworkRetries = 12
)

func main() {
	cfg, err := parseConfig()
	if err != nil {
		panic(fmt.Errorf("error parsing command-line arguments: %s", err))
	}

	privateKey, addrPubKeyHash, err := decodeKeys(cfg)
	if err != nil {
		panic(fmt.Errorf("error decoding public key: %s", err))
	}

	client, err := connect(cfg)
	if err != nil {
		panic(fmt.Errorf("could not connect to RPC server: %s", err))
	}
	log.Infof("Connected to server %s", cfg.RPCServer)

	fundingOutPoint, fundingTx, err := findUnspentTXO(client, addrPubKeyHash)
	if err != nil {
		panic(fmt.Errorf("error finding unspent transactions: %s", err))
	}
	if fundingOutPoint == nil || fundingTx == nil {
		panic(fmt.Errorf("could not find any unspent transactions this for key"))
	}
	log.Infof("Found transaction to spend: %s:%d", fundingOutPoint.TxID, fundingOutPoint.Index)

	registryTx, err := buildSubnetworkRegistryTx(cfg, fundingOutPoint, fundingTx, privateKey)
	if err != nil {
		panic(fmt.Errorf("error building subnetwork registry tx: %s", err))
	}

	_, err = client.SendRawTransaction(registryTx, true)
	if err != nil {
		panic(fmt.Errorf("failed sending subnetwork registry tx: %s", err))
	}
	log.Infof("Successfully sent subnetwork registry transaction")

	subnetworkID, err := blockdag.TxToSubnetworkID(registryTx)
	if err != nil {
		panic(fmt.Errorf("could not build subnetwork ID: %s", err))
	}

	err = waitForSubnetworkToBecomeAccepted(client, subnetworkID)
	if err != nil {
		panic(fmt.Errorf("error waiting for subnetwork to become accepted: %s", err))
	}
	log.Infof("Subnetwork '%s' was successfully registered.", subnetworkID)
}

func waitForSubnetworkToBecomeAccepted(client *rpcclient.Client, subnetworkID *subnetworkid.SubnetworkID) error {
	retries := 0
	for {
		_, err := client.GetSubnetwork(subnetworkID.String())
		if err != nil {
			if rpcError, ok := err.(*btcjson.RPCError); ok && rpcError.Code == btcjson.ErrRPCSubnetworkNotFound {
				log.Infof("Subnetwork not found")

				retries++
				if retries == maxGetSubnetworkRetries {
					return fmt.Errorf("failed to get subnetwork %d times: %s", maxGetSubnetworkRetries, err)
				}

				log.Infof("Waiting %d seconds...", int(getSubnetworkRetryDelay.Seconds()))
				<-time.After(getSubnetworkRetryDelay)
				continue
			}
			return fmt.Errorf("failed getting subnetwork: %s", err)
		}
		return nil
	}
}
