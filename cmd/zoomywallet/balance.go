package main

import (
	"context"
	"fmt"

	"github.com/zoomy-network/zoomyd/cmd/zoomywallet/daemon/client"
	"github.com/zoomy-network/zoomyd/cmd/zoomywallet/daemon/pb"
	"github.com/zoomy-network/zoomyd/cmd/zoomywallet/utils"
)

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

	pendingSuffix := ""
	if response.Pending > 0 {
		pendingSuffix = " (pending)"
	}
	if conf.Verbose {
		pendingSuffix = ""
		println("Address                                                                       Available             Pending")
		println("-----------------------------------------------------------------------------------------------------------")
		for _, addressBalance := range response.AddressBalances {
			fmt.Printf("%s %s %s\n", addressBalance.Address, utils.FormatKas(addressBalance.Available), utils.FormatKas(addressBalance.Pending))
		}
		println("-----------------------------------------------------------------------------------------------------------")
		print("                                                 ")
	}
	fmt.Printf("Total balance, ARF %s %s%s\n", utils.FormatKas(response.Available), utils.FormatKas(response.Pending), pendingSuffix)

	return nil
}
