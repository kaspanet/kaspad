package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"os"
)

func isUTXOSpendable(entry *appmessage.UTXOsByAddressesEntry, virtualSelectedParentBlueScore uint64, coinbaseMaturity uint64) bool {
	if !entry.UTXOEntry.IsCoinbase {
		return true
	}
	// TODO: DECIDE HOW TO HANDLE COINBASE MATURITY
	blockBlueScore := entry.UTXOEntry.BlockBlueScore
	return blockBlueScore+coinbaseMaturity < virtualSelectedParentBlueScore
}

func printErrorAndExit(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}
