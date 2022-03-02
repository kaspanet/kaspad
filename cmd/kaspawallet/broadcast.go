package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/client"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/pkg/errors"
)

func broadcast(conf *broadcastConfig) error {
	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	if conf.Transaction == "" && conf.TransactionFile == "" {
		return errors.Errorf("Either --transaction or --transaction-file is required")
	}
	if conf.Transaction != "" && conf.TransactionFile != "" {
		return errors.Errorf("Both --transaction and --transaction-file cannot be passed at the same time")
	}

	transactionsHex := conf.Transaction
	if conf.TransactionFile != "" {
		transactionHexBytes, err := ioutil.ReadFile(conf.TransactionFile)
		if err != nil {
			return errors.Wrapf(err, "Could not read hex from %s", conf.TransactionFile)
		}
		transactionsHex = strings.TrimSpace(string(transactionHexBytes))
	}

	transactions, err := decodeTransactionsFromHex(transactionsHex)
	if err != nil {
		return err
	}

	for _, transaction := range transactions {
		response, err := daemonClient.Broadcast(ctx, &pb.BroadcastRequest{Transaction: transaction})
		if err != nil {
			return err
		}
		fmt.Println("Transaction was sent successfully")
		fmt.Printf("Transaction ID: \t%s\n", response.TxID)
	}

	return nil
}
