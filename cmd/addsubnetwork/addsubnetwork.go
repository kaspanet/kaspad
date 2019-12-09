package main

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/kaspajson"
	"github.com/kaspanet/kaspad/rpcclient"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/pkg/errors"
	"time"
)

const (
	getSubnetworkRetryDelay = 5 * time.Second
	maxGetSubnetworkRetries = 12
)

func main() {
	cfg, err := parseConfig()
	if err != nil {
		panic(errors.Errorf("error parsing command-line arguments: %s", err))
	}

	privateKey, addrPubKeyHash, err := decodeKeys(cfg)
	if err != nil {
		panic(errors.Errorf("error decoding public key: %s", err))
	}

	client, err := connect(cfg)
	if err != nil {
		panic(errors.Errorf("could not connect to RPC server: %s", err))
	}
	log.Infof("Connected to server %s", cfg.RPCServer)

	fundingOutpoint, fundingTx, err := findUnspentTXO(cfg, client, addrPubKeyHash)
	if err != nil {
		panic(errors.Errorf("error finding unspent transactions: %s", err))
	}
	if fundingOutpoint == nil || fundingTx == nil {
		panic(errors.Errorf("could not find any unspent transactions this for key"))
	}
	log.Infof("Found transaction to spend: %s:%d", fundingOutpoint.TxID, fundingOutpoint.Index)

	registryTx, err := buildSubnetworkRegistryTx(cfg, fundingOutpoint, fundingTx, privateKey)
	if err != nil {
		panic(errors.Errorf("error building subnetwork registry tx: %s", err))
	}

	_, err = client.SendRawTransaction(registryTx, true)
	if err != nil {
		panic(errors.Errorf("failed sending subnetwork registry tx: %s", err))
	}
	log.Infof("Successfully sent subnetwork registry transaction")

	subnetworkID, err := blockdag.TxToSubnetworkID(registryTx)
	if err != nil {
		panic(errors.Errorf("could not build subnetwork ID: %s", err))
	}

	err = waitForSubnetworkToBecomeAccepted(client, subnetworkID)
	if err != nil {
		panic(errors.Errorf("error waiting for subnetwork to become accepted: %s", err))
	}
	log.Infof("Subnetwork '%s' was successfully registered.", subnetworkID)
}

func waitForSubnetworkToBecomeAccepted(client *rpcclient.Client, subnetworkID *subnetworkid.SubnetworkID) error {
	retries := 0
	for {
		_, err := client.GetSubnetwork(subnetworkID.String())
		if err != nil {
			if rpcError, ok := err.(*kaspajson.RPCError); ok && rpcError.Code == kaspajson.ErrRPCSubnetworkNotFound {
				log.Infof("Subnetwork not found")

				retries++
				if retries == maxGetSubnetworkRetries {
					return errors.Errorf("failed to get subnetwork %d times: %s", maxGetSubnetworkRetries, err)
				}

				log.Infof("Waiting %d seconds...", int(getSubnetworkRetryDelay.Seconds()))
				<-time.After(getSubnetworkRetryDelay)
				continue
			}
			return errors.Errorf("failed getting subnetwork: %s", err)
		}
		return nil
	}
}
