package main

import (
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
)

func broadcast(conf *broadcastConfig) error {
	client, err := rpcclient.NewRPCClient(conf.RPCServer)
	if err != nil {
		return err
	}

	psTxBytes, err := hex.DecodeString(conf.Transaction)
	if err != nil {
		return err
	}

	tx, err := libkaspawallet.ExtractTransaction(psTxBytes)
	if err != nil {
		return err
	}

	transactionID, err := sendTransaction(client, tx)
	if err != nil {
		return err
	}

	fmt.Println("Transaction was sent successfully")
	fmt.Printf("Transaction ID: \t%s\n", transactionID)

	return nil
}
