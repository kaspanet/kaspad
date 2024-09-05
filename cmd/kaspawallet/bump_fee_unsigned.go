package main

import (
	"context"
	"fmt"
	"math"
	"os"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/client"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
)

func bumpFeeUnsigned(conf *bumpFeeUnsignedConfig) error {
	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	if err != nil {
		return err
	}

	feeRate := &pb.FeeRate{
		FeeRate: &pb.FeeRate_Max{Max: math.MaxFloat64},
	}
	if conf.FeeRate > 0 {
		feeRate.FeeRate = &pb.FeeRate_Exact{Exact: conf.FeeRate}
	} else if conf.MaxFeeRate > 0 {
		feeRate.FeeRate = &pb.FeeRate_Max{Max: conf.MaxFeeRate}
	}

	response, err := daemonClient.BumpFee(ctx, &pb.BumpFeeRequest{
		TxID:                     conf.TxID,
		From:                     conf.FromAddresses,
		UseExistingChangeAddress: conf.UseExistingChangeAddress,
		FeeRate:                  feeRate,
	})
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "Created unsigned transaction")
	fmt.Println(encodeTransactionsToHex(response.Transactions))

	return nil
}
