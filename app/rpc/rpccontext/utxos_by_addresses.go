package rpccontext

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/utxoindex"
)

func ConvertUTXOOutpointEntryPairsToUTXOsByAddressesEntries(address string, pairs utxoindex.UTXOOutpointEntryPairs) []*appmessage.UTXOsByAddressesEntry {
	utxosByAddressesEntries := make([]*appmessage.UTXOsByAddressesEntry, 0, len(pairs))
	for outpoint, utxoEntry := range pairs {
		utxosByAddressesEntries = append(utxosByAddressesEntries, &appmessage.UTXOsByAddressesEntry{
			Address: address,
			Outpoint: &appmessage.RPCOutpoint{
				TransactionID: hex.EncodeToString(outpoint.TransactionID[:]),
				Index:         outpoint.Index,
			},
			UTXOEntry: &appmessage.RPCUTXOEntry{
				Amount:         utxoEntry.Amount(),
				ScriptPubKey:   hex.EncodeToString(utxoEntry.ScriptPublicKey()),
				BlockBlueScore: utxoEntry.BlockBlueScore(),
				IsCoinbase:     utxoEntry.IsCoinbase(),
			},
		})
	}
	return utxosByAddressesEntries
}
