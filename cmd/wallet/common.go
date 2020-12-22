package main

import "github.com/kaspanet/kaspad/app/appmessage"

const minConfirmations = 100

func isUTXOSpendable(entry *appmessage.UTXOsByAddressesEntry, virtualSelectedParentBlueScore uint64) bool {
	if !entry.UTXOEntry.IsCoinbase {
		return true
	}
	blockBlueScore := entry.UTXOEntry.BlockBlueScore
	return blockBlueScore+minConfirmations < virtualSelectedParentBlueScore
}
