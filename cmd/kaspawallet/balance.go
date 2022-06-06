package main

import (
	"context"
	"fmt"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/client"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/utils"
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
		if conf.CountUtxos {
			println("Address                                                                       Available   #UTXOs             Pending   #UTXOs")
			println("-----------------------------------------------------------------------------------------------------------------------------")
			for _, addressBalance := range response.AddressBalances {
				fmt.Printf("%s %s %s %s %s\n", addressBalance.Address, utils.FormatKas(addressBalance.Available),
					utils.FormatUtxos(addressBalance.NUtxosAvailable), utils.FormatKas(addressBalance.Pending),
					utils.FormatUtxos(addressBalance.NUtxosPending))
			}
			println("-----------------------------------------------------------------------------------------------------------------------------")
		} else {
			println("Address                                                                       Available             Pending")
			println("-----------------------------------------------------------------------------------------------------------")
			for _, addressBalance := range response.AddressBalances {
				fmt.Printf("%s %s %s\n", addressBalance.Address, utils.FormatKas(addressBalance.Available), utils.FormatKas(addressBalance.Pending))
			}
			println("-----------------------------------------------------------------------------------------------------------")
		}
		print("                                                 ")
	}
	if conf.CountUtxos {
		if response.Pending > 0 {
			fmt.Printf("Total balance, KAS %s (%d UTXOs) %s pending (%d UTXOs)\n", utils.FormatKas(response.Available),
				response.NUtxosAvailable, utils.FormatKas(response.Pending), response.NUtxosPending)
		} else {
			fmt.Printf("Total balance, KAS %s (%d UTXOs)\n", utils.FormatKas(response.Available), response.NUtxosAvailable)
		}
	} else {
		fmt.Printf("Total balance, KAS %s %s%s\n", utils.FormatKas(response.Available), utils.FormatKas(response.Pending), pendingSuffix)
	}

	return nil
}
