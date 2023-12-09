package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/cmd/zoomywallet/daemon/client"
	"github.com/zoomy-network/zoomyd/cmd/zoomywallet/daemon/pb"
)

func broadcast(conf *broadcastConfig) error {
	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	if conf.Transactions == "" && conf.TransactionsFile == "" {
		return errors.Errorf("Either --transaction or --transaction-file is required")
	}
	if conf.Transactions != "" && conf.TransactionsFile != "" {
		return errors.Errorf("Both --transaction and --transaction-file cannot be passed at the same time")
	}

	transactionsHex := conf.Transactions
	if conf.TransactionsFile != "" {
		transactionHexBytes, err := ioutil.ReadFile(conf.TransactionsFile)
		if err != nil {
			return errors.Wrapf(err, "Could not read hex from %s", conf.TransactionsFile)
		}
		transactionsHex = strings.TrimSpace(string(transactionHexBytes))
	}

	transactions, err := decodeTransactionsFromHex(transactionsHex)
	if err != nil {
		return err
	}

	response, err := daemonClient.Broadcast(ctx, &pb.BroadcastRequest{Transactions: transactions})
	if err != nil {
		return err
	}
	fmt.Println("Transactions were sent successfully")
	fmt.Println("Transaction ID(s): ")
	for _, txID := range response.TxIDs {
		fmt.Printf("\t%s\n", txID)
	}

	return nil
}
