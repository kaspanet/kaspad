package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"os"
)

const minConfirmations = 100

func isUTXOSpendable(entry *appmessage.UTXOsByAddressesEntry, virtualSelectedParentBlueScore uint64) bool {
	if !entry.UTXOEntry.IsCoinbase {
		return true
	}
	blockBlueScore := entry.UTXOEntry.BlockBlueScore
	return blockBlueScore+minConfirmations < virtualSelectedParentBlueScore
}

func printErrorAndExit(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}
