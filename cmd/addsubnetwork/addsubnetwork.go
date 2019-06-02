package main

import (
	"fmt"
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/rpcclient"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util/subnetworkid"
	"github.com/daglabs/btcd/wire"
	"log"
	"time"
)

const (
	newSubnetworkGasLimit   = 1000
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

	fundingOutPoint, fundingTx, err := findUnspentTXO(client, addrPubKeyHash)
	if err != nil {
		panic(fmt.Errorf("error finding unspent transactions: %s", err))
	}
	if fundingOutPoint == nil || fundingTx == nil {
		panic(fmt.Errorf("could not find any unspent transactions this for key"))
	}
	log.Printf("found transaction to spend: %s:%d", fundingOutPoint.TxID, fundingOutPoint.Index)

	registryTx, err := buildSubnetworkRegistryTx(fundingOutPoint, fundingTx, privateKey)
	if err != nil {
		panic(fmt.Errorf("error building subnetwork registry tx: %s", err))
	}

	_, err = client.SendRawTransaction(registryTx, true)
	if err != nil {
		panic(fmt.Errorf("failed sending subnetwork registry tx: %s", err))
	}
	log.Printf("successfully sent subnetwork registry transaction")

	subnetworkID, err := blockdag.TxToSubnetworkID(registryTx)
	if err != nil {
		panic(fmt.Errorf("could not build subnetwork ID: %s", err))
	}

	wasAccepted, err := waitForSubnetworkToBecomeAccepted(client, subnetworkID)
	if err != nil {
		panic(fmt.Errorf("error waiting for subnetwork to become accepted: %s", err))
	}

	if wasAccepted {
		log.Printf("Subnetwork '%s' was successfully registered.", subnetworkID)
	} else {
		log.Printf("Subnetwork '%s' did not register.", subnetworkID)
	}
}

func buildSubnetworkRegistryTx(fundingOutPoint *wire.OutPoint, fundingTx *wire.MsgTx, privateKey *btcec.PrivateKey) (*wire.MsgTx, error) {
	txIn := &wire.TxIn{
		PreviousOutPoint: *fundingOutPoint,
		Sequence:         wire.MaxTxInSequenceNum,
	}
	pkScript, err := txscript.PayToScriptHashScript(blockdag.OpTrueScript)
	if err != nil {
		return nil, err
	}
	txOut := &wire.TxOut{
		PkScript: pkScript,
		Value:    fundingTx.TxOut[fundingOutPoint.Index].Value,
	}
	registryTx := wire.NewRegistryMsgTx(1, []*wire.TxIn{txIn}, []*wire.TxOut{txOut}, newSubnetworkGasLimit)

	SignatureScript, err := txscript.SignatureScript(registryTx, 0, fundingTx.TxOut[fundingOutPoint.Index].PkScript,
		txscript.SigHashAll, privateKey, true)
	if err != nil {
		return nil, fmt.Errorf("failed to build signature script: %s", err)
	}
	txIn.SignatureScript = SignatureScript

	return registryTx, nil
}

func waitForSubnetworkToBecomeAccepted(client *rpcclient.Client, subnetworkID *subnetworkid.SubnetworkID) (bool, error) {
	retries := 0
	for {
		_, err := client.GetSubnetwork(subnetworkID.String())
		if err != nil {
			if rpcError, ok := err.(*btcjson.RPCError); ok && rpcError.Code == btcjson.ErrRPCSubnetworkNotFound {
				log.Printf("Subnetwork not found")

				retries++
				if retries == maxGetSubnetworkRetries {
					return false, nil
				}

				log.Printf("Waiting %d seconds...", int(getSubnetworkRetryDelay.Seconds()))
				<-time.After(getSubnetworkRetryDelay)
				continue
			}
			return false, fmt.Errorf("failed getting subnetwork: %s", err)
		}
		return true, nil
	}
}
