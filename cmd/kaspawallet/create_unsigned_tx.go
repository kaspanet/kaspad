package main

import (
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/util"
)

func createUnsignedTransaction(conf *createUnsignedTransactionConfig) error {
	toAddress, err := util.DecodeAddress(conf.ToAddress, conf.NetParams().Prefix)
	if err != nil {
		return err
	}

	pubKeys := make([][]byte, len(conf.PublicKey))
	for i, pubKeyHex := range conf.PublicKey {
		pubKeys[i], err = hex.DecodeString(pubKeyHex)
		if err != nil {
			return err
		}
	}

	fromAddress, err := libkaspawallet.MultiSigAddress(conf.NetParams(), pubKeys, conf.MinimumSignatures)
	if err != nil {
		return err
	}

	client, err := rpcclient.NewRPCClient(conf.RPCServer)
	if err != nil {
		return err
	}
	utxos, err := fetchSpendableUTXOs(conf.NetParams(), client, fromAddress.String())
	if err != nil {
		return err
	}

	sendAmountSompi := uint64(conf.SendAmount * util.SompiPerKaspa)

	const feePerInput = 1000
	selectedUTXOs, changeSompi, err := selectUTXOs(utxos, sendAmountSompi, feePerInput)
	if err != nil {
		return err
	}

	psTx, err := libkaspawallet.CreateUnsignedTransaction(pubKeys, conf.MinimumSignatures, []*libkaspawallet.Payment{{
		Address: toAddress,
		Amount:  sendAmountSompi,
	}, {
		Address: fromAddress,
		Amount:  changeSompi,
	}}, selectedUTXOs)
	if err != nil {
		return err
	}

	fmt.Println("Created unsigned transaction")
	fmt.Println(hex.EncodeToString(psTx))
	return nil
}
