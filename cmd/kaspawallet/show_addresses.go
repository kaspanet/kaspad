package main

import (
	"context"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/client"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
)

func showAddresses(conf *showAddressesConfig) error {
	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	response, err := daemonClient.ShowAddresses(ctx, &pb.ShowAddressesRequest{})
	if err != nil {
		return err
	}

	fmt.Printf("Addresses (%d):\n", len(response.Address))
	for _, address := range response.Address {
		fmt.Println(address)
	}
	return nil
}
