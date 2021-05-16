package main

import (
	"context"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/client"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/util"
)

const (
	externalKeychain = 0
	internalKeychain = 1
)

var keychains = []uint8{externalKeychain, internalKeychain}

func balance(conf *balanceConfig) error {
	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()
	response, err := daemonClient.GetBalance(ctx, &pb.GetBalanceRequest{})
	if err != nil {
		return err
	}

	fmt.Printf("Balance:\t\tKAS %f\n", float64(response.Available)/util.SompiPerKaspa)
	if response.Pending > 0 {
		fmt.Printf("Pending balance:\tKAS %f\n", float64(response.Pending)/util.SompiPerKaspa)
	}

	return nil
}
