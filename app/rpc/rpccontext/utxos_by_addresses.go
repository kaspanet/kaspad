package rpccontext

import (
	"encoding/hex"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/utxoindex"
)

// ConvertDomainOutpointEntryPairsToUTXOsByAddressesEntries converts
// *externalapi.OutpointAndUTXOEntryPairs to a slice of UTXOsByAddressesEntry
func ConvertDomainOutpointEntryPairsToUTXOsByAddressesEntries(address string, pairs []*externalapi.OutpointAndUTXOEntryPair) []*appmessage.UTXOsByAddressesEntry {
	utxosByAddressesEntries := make([]*appmessage.UTXOsByAddressesEntry, 0, len(pairs))
	for _, outpointAndUTXOEntryPair := range pairs {
		utxosByAddressesEntries = append(utxosByAddressesEntries, &appmessage.UTXOsByAddressesEntry{
			Address: address,
			Outpoint: &appmessage.RPCOutpoint{
				TransactionID: outpointAndUTXOEntryPair.Outpoint.TransactionID.String(),
				Index:         outpointAndUTXOEntryPair.Outpoint.Index,
			},
			UTXOEntry: &appmessage.RPCUTXOEntry{
				Amount: outpointAndUTXOEntryPair.UTXOEntry.Amount(),
				ScriptPublicKey: &appmessage.RPCScriptPublicKey{
					Script:  hex.EncodeToString(outpointAndUTXOEntryPair.UTXOEntry.ScriptPublicKey().Script),
					Version: outpointAndUTXOEntryPair.UTXOEntry.ScriptPublicKey().Version},
				BlockDAAScore: outpointAndUTXOEntryPair.UTXOEntry.BlockDAAScore(),
				IsCoinbase:    outpointAndUTXOEntryPair.UTXOEntry.IsCoinbase(),
			},
		})
	}
	return utxosByAddressesEntries
}

// convertUTXOOutpointsToUTXOsByAddressesEntries converts
// UTXOOutpoints to a slice of UTXOsByAddressesEntry
func convertUTXOOutpointsToUTXOsByAddressesEntries(address string, outpoints utxoindex.UTXOOutpoints) []*appmessage.UTXOsByAddressesEntry {
	utxosByAddressesEntries := make([]*appmessage.UTXOsByAddressesEntry, 0, len(outpoints))
	for outpoint := range outpoints {
		utxosByAddressesEntries = append(utxosByAddressesEntries, &appmessage.UTXOsByAddressesEntry{
			Address: address,
			Outpoint: &appmessage.RPCOutpoint{
				TransactionID: outpoint.TransactionID.String(),
				Index:         outpoint.Index,
			},
		})
	}
	return utxosByAddressesEntries
}

// ConvertAddressStringsToUTXOsChangedNotificationAddresses converts address strings
// to UTXOsChangedNotificationAddresses
func (ctx *Context) ConvertAddressStringsToUTXOsChangedNotificationAddresses(
	addressStrings []string) ([]*UTXOsChangedNotificationAddress, error) {

	addresses := make([]*UTXOsChangedNotificationAddress, len(addressStrings))
	for i, addressString := range addressStrings {
		address, err := util.DecodeAddress(addressString, ctx.Config.ActiveNetParams.Prefix)
		if err != nil {
			return nil, errors.Errorf("Could not decode address '%s': %s", addressString, err)
		}
		scriptPublicKey, err := txscript.PayToAddrScript(address)
		if err != nil {
			return nil, errors.Errorf("Could not create a scriptPublicKey for address '%s': %s", addressString, err)
		}
		scriptPublicKeyString := utxoindex.ConvertScriptPublicKeyToString(scriptPublicKey)
		addresses[i] = &UTXOsChangedNotificationAddress{
			Address:               addressString,
			ScriptPublicKeyString: scriptPublicKeyString,
		}
	}
	return addresses, nil
}
