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

	feePolicy := &pb.FeePolicy{
		FeePolicy: &pb.FeePolicy_MaxFeeRate{MaxFeeRate: math.MaxFloat64},
	}
	if conf.FeeRate > 0 {
		feePolicy.FeePolicy = &pb.FeePolicy_ExactFeeRate{ExactFeeRate: conf.FeeRate}
	} else if conf.MaxFeeRate > 0 {
		feePolicy.FeePolicy = &pb.FeePolicy_MaxFeeRate{MaxFeeRate: conf.MaxFeeRate}
	}

	response, err := daemonClient.BumpFee(ctx, &pb.BumpFeeRequest{
		TxID:                     conf.TxID,
		From:                     conf.FromAddresses,
		UseExistingChangeAddress: conf.UseExistingChangeAddress,
		FeePolicy:                  feePolicy,
	})
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "Created unsigned transaction")
	fmt.Println(encodeTransactionsToHex(response.Transactions))

	return nil
}