package main

import (
	"context"
	"fmt"

	"github.com/c4ei/YunSeokYeol/cmd/kaspawallet/daemon/client"
	"github.com/c4ei/YunSeokYeol/cmd/kaspawallet/daemon/pb"
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

	fmt.Printf("\nNote: the above are only addresses that were manually created by the 'new-address' command. If you want to see a list of all addresses, including change addresses, " +
		"that have a positive balance, use the command 'balance -v'\n")
	return nil
}
