package main

import (
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/keys"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/util"
)

func createUnsignedTransaction(conf *createUnsignedTransactionConfig) error {
	keysFile, err := keys.ReadKeysFile(conf.KeysFile)
	if err != nil {
		return err
	}

	toAddress, err := util.DecodeAddress(conf.ToAddress, conf.NetParams().Prefix)
	if err != nil {
		return err
	}

	fromAddress, err := libkaspawallet.Address(conf.NetParams(), keysFile.PublicKeys, keysFile.MinimumSignatures)
	if err != nil {
		return err
	}

	client, err := connectToRPC(conf.NetParams(), conf.RPCServer)
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

	psTx, err := libkaspawallet.CreateUnsignedTransaction(keysFile.PublicKeys, keysFile.MinimumSignatures, []*libkaspawallet.Payment{{
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
