package rpccontext

import (
	"encoding/hex"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/utxoindex"
)

// ConvertUTXOOutpointEntryPairsToUTXOsByAddressesEntries converts
// UTXOOutpointEntryPairs to a slice of UTXOsByAddressesEntry
func ConvertUTXOOutpointEntryPairsToUTXOsByAddressesEntries(address string, pairs utxoindex.UTXOOutpointEntryPairs) []*appmessage.UTXOsByAddressesEntry {
	utxosByAddressesEntries := make([]*appmessage.UTXOsByAddressesEntry, 0, len(pairs))
	for outpoint, utxoEntry := range pairs {
		utxosByAddressesEntries = append(utxosByAddressesEntries, &appmessage.UTXOsByAddressesEntry{
			Address: address,
			Outpoint: &appmessage.RPCOutpoint{
				TransactionID: outpoint.TransactionID.String(),
				Index:         outpoint.Index,
			},
			UTXOEntry: &appmessage.RPCUTXOEntry{
				Amount:         utxoEntry.Amount(),
				ScriptPubKey:   &appmessage.RPCScriptPubKey{Script: hex.EncodeToString(utxoEntry.ScriptPublicKey().Script), Version: utxoEntry.ScriptPublicKey().Version},
				BlockBlueScore: utxoEntry.BlockBlueScore(),
				IsCoinbase:     utxoEntry.IsCoinbase(),
			},
		})
	}
	return utxosByAddressesEntries
}
