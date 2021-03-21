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
	blockBlueScore := entry.UTXOEntry.BlockDAAScore
	// TODO: Check for a better alternative than virtualSelectedParentBlueScore
	return blockBlueScore+coinbaseMaturity < virtualSelectedParentBlueScore
}

func printErrorAndExit(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}
