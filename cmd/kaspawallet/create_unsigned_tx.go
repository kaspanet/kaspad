package main

import (
	"context"
	"fmt"
	"math"
	"os"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/client"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/utils"
)

func createUnsignedTransaction(conf *createUnsignedTransactionConfig) error {
	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	var sendAmountSompi uint64

	if !conf.IsSendAll {
		sendAmountSompi, err = utils.KasToSompi(conf.SendAmount)
		if err != nil {
			return err
		}
	}

	feeRate := &pb.FeeRate{
		FeeRate: &pb.FeeRate_Max{Max: math.MaxFloat64},
	}
	if conf.FeeRate > 0 {
		feeRate.FeeRate = &pb.FeeRate_Exact{Exact: conf.FeeRate}
	} else if conf.MaxFeeRate > 0 {
		feeRate.FeeRate = &pb.FeeRate_Max{Max: conf.MaxFeeRate}
	}

	response, err := daemonClient.CreateUnsignedTransactions(ctx, &pb.CreateUnsignedTransactionsRequest{
		From:                     conf.FromAddresses,
		Address:                  conf.ToAddress,
		Amount:                   sendAmountSompi,
		IsSendAll:                conf.IsSendAll,
		UseExistingChangeAddress: conf.UseExistingChangeAddress,
		FeeRate:                  feeRate,
	})
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "Created unsigned transaction")
	fmt.Println(encodeTransactionsToHex(response.UnsignedTransactions))

	return nil
}
