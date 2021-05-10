package main

import (
	"context"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/client"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
)

func showAddress(conf *showAddressConfig) error {
	daemonClient, tearDown, err := client.Connect(conf.ServerAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	response, err := daemonClient.GetReceiveAddress(ctx, &pb.GetReceiveAddressRequest{})
	if err != nil {
		return err
	}

	fmt.Printf("Address:\n%s\n", response.Address)
	return nil
}
