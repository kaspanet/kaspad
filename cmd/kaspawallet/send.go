package main

import (
	"context"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/client"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/keys"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/util"
)

func send(conf *sendConfig) error {
	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	sendAmountSompi := uint64(conf.SendAmount * util.SompiPerKaspa)
	createUnsignedTransactionResponse, err := daemonClient.CreateUnsignedTransaction(ctx, &pb.CreateUnsignedTransactionRequest{
		Address: conf.ToAddress,
		Amount:  sendAmountSompi,
	})
	if err != nil {
		return err
	}

	keysFile, err := keys.ReadKeysFile(conf.NetParams(), conf.KeysFile)
	if err != nil {
		return err
	}

	mnemonics, err := keysFile.DecryptMnemonics()
	if err != nil {
		return err
	}

	updatedPSTx, err := libkaspawallet.Sign(conf.NetParams(), mnemonics, createUnsignedTransactionResponse.UnsignedTransaction, keysFile.ECDSA)
	if err != nil {
		return err
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel2()
	broadcastResponse, err := daemonClient.Broadcast(ctx2, &pb.BroadcastRequest{
		Transaction: updatedPSTx,
	})
	if err != nil {
		return err
	}

	fmt.Println("Transaction was sent successfully")
	fmt.Printf("Transaction ID: \t%s\n", broadcastResponse.TxID)

	return nil
}
