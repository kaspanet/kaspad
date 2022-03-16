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

	transactionsCount := len(transactions)
	for i, transaction := range transactions {
		response, err := daemonClient.Broadcast(ctx, &pb.BroadcastRequest{Transaction: transaction})
		if err != nil {
			return err
		}
		if transactionsCount == 1 {
			fmt.Println("Transactions was sent successfully")
		} else {
			fmt.Printf("Transactions %d (out of %d) was sent successfully\n", i+1, transactionsCount)
		}
		fmt.Printf("Transactions ID: \t%s\n", response.TxID)
	}

	return nil
}
