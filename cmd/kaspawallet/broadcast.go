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

	const batch = 50
	position := 0

	// We send in batches to avoid communication errors
	for position < len(transactions) {
		var transactionBatch [][]byte
		if position+batch > len(transactions) {
			transactionBatch = transactions[position:]
		} else {
			transactionBatch = transactions[position : position+batch]
		}
		response, err := daemonClient.Broadcast(ctx, &pb.BroadcastRequest{Transactions: transactionBatch})
		if err != nil {
			return err
		}
		if position == 0 {
			fmt.Println("Transactions were sent successfully")
		}
		fmt.Println("Transaction ID(s): ")
		for _, txID := range response.TxIDs {
			fmt.Printf("\t%s\n", txID)
		}
		position += batch
	}

	return nil
}
