package main

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/client"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
)

func createUnsignedTransaction(conf *createUnsignedTransactionConfig) error {
	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	sendAmountSompi := uint64(conf.SendAmount * constants.SompiPerKaspa)
	response, err := daemonClient.CreateUnsignedTransactions(ctx, &pb.CreateUnsignedTransactionsRequest{
		Address: conf.ToAddress,
		Amount:  sendAmountSompi,
	})
	if err != nil {
		return err
	}

	fmt.Println("Created unsigned transaction")
	for _, unsignedTransaction := range response.UnsignedTransactions {
		fmt.Println(hex.EncodeToString(unsignedTransaction))
	}
	return nil
}
