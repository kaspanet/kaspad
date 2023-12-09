package main

import (
	"context"
	"fmt"
<<<<<<< Updated upstream:cmd/kaspawallet/new_address.go
=======

>>>>>>> Stashed changes:cmd/zoomywallet/new_address.go
	"github.com/zoomy-network/zoomyd/cmd/zoomywallet/daemon/client"
	"github.com/zoomy-network/zoomyd/cmd/zoomywallet/daemon/pb"
)

func newAddress(conf *newAddressConfig) error {
	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	response, err := daemonClient.NewAddress(ctx, &pb.NewAddressRequest{})
	if err != nil {
		return err
	}

	fmt.Printf("New address:\n%s\n", response.Address)
	return nil
}
