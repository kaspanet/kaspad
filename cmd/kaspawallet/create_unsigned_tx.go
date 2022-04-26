package main

import (
	"context"
	"fmt"
	"strings"

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

	var fromAddresses []string
	if conf.FromAddresses != "" {
		fromAddresses = strings.Split(conf.FromAddresses, ",")
	}

	sendAmountSompi := uint64(conf.SendAmount * constants.SompiPerKaspa)
	response, err := daemonClient.CreateUnsignedTransactions(ctx, &pb.CreateUnsignedTransactionsRequest{
		From:    fromAddresses,
		Address: conf.ToAddress,
		Amount:  sendAmountSompi,
	})
	if err != nil {
		return err
	}

	fmt.Println("Created unsigned transaction")
	fmt.Println(encodeTransactionsToHex(response.UnsignedTransactions))
	return nil
}
