package main

import (
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
)

func broadcast(conf *broadcastConfig) error {
	client, err := connectToRPC(conf.NetParams(), conf.RPCServer)
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
