package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/client"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
)

func broadcast(conf *broadcastConfig) error {
	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	transaction, err := hex.DecodeString(conf.Transaction)
	if err != nil {
		return err
	}

	response, err := daemonClient.Broadcast(ctx, &pb.BroadcastRequest{Transaction: transaction})
	if err != nil {
		return err
	}

	fmt.Println("Transaction was sent successfully")
	fmt.Printf("Transaction ID: \t%s\n", response.TxID)

	return nil
}
