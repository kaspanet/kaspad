package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/domain/utxoindex"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util"
	"sort"
)

// HandleGetUTXOsByAddresses handles the respectively named RPC command
func HandleGetUTXOsByAddresses(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	if !context.Config.UTXOIndex {
		errorMessage := &appmessage.GetUTXOsByAddressesResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Method unavailable when kaspad is run without --utxoindex")
		return errorMessage, nil
	}

	getUTXOsByAddressesRequest := request.(*appmessage.GetUTXOsByAddressesRequestMessage)

	allEntries := make([]*appmessage.UTXOsByAddressesEntry, 0)
	for _, addressString := range getUTXOsByAddressesRequest.Addresses {
		address, err := util.DecodeAddress(addressString, context.Config.ActiveNetParams.Prefix)
		if err != nil {
			errorMessage := &appmessage.GetUTXOsByAddressesResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("Could not decode address '%s': %s", addressString, err)
			return errorMessage, nil
		}
		scriptPublicKey, err := txscript.PayToAddrScript(address)
		if err != nil {
			errorMessage := &appmessage.GetUTXOsByAddressesResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("Could not create a scriptPublicKey for address '%s': %s", addressString, err)
			return errorMessage, nil
		}
		utxoOutpointEntryPairs, err := context.UTXOIndex.UTXOs(scriptPublicKey)
		if err != nil {
			return nil, err
		}

		if getUTXOsByAddressesRequest.BatchDaaScoreStart > 0 ||
			(getUTXOsByAddressesRequest.BatchSize > 0 && uint64(len(utxoOutpointEntryPairs)) > getUTXOsByAddressesRequest.BatchSize) {
			// Find a batch of entries with consecutive DAA score
			entriesOrderedBatch := extractOrderedEntriesBatch(utxoOutpointEntryPairs, getUTXOsByAddressesRequest.BatchDaaScoreStart, getUTXOsByAddressesRequest.BatchSize)
			// Extract the batch from the full pairs map
			entries := rpccontext.ConvertUTXOOutpointEntryBatchToUTXOsByAddressesEntries(addressString, utxoOutpointEntryPairs, entriesOrderedBatch)
			allEntries = append(allEntries, entries...)
		} else {
			entries := rpccontext.ConvertUTXOOutpointEntryPairsToUTXOsByAddressesEntries(addressString, utxoOutpointEntryPairs)
			allEntries = append(allEntries, entries...)
		}
	}

	response := appmessage.NewGetUTXOsByAddressesResponseMessage(allEntries)
	return response, nil
}

func extractOrderedEntriesBatch(utxoOutpointEntryPairs utxoindex.UTXOOutpointEntryPairs, batchDaaScoreStart, batchSize uint64) []rpccontext.OutpointDAAScoreEntry {
	entriesSlice := make([]rpccontext.OutpointDAAScoreEntry, 0, len(utxoOutpointEntryPairs))
	// Extract to slice
	for outpoint, utxoEntry := range utxoOutpointEntryPairs {
		entriesSlice = append(entriesSlice, rpccontext.OutpointDAAScoreEntry{DAAScore: utxoEntry.BlockDAAScore(), Outpoint: outpoint})
	}
	// Sort by DAA score
	sort.Slice(entriesSlice, func(i, j int) bool {
		if entriesSlice[i].DAAScore == entriesSlice[j].DAAScore {
			if entriesSlice[i].Outpoint.TransactionID.Equal(&entriesSlice[j].Outpoint.TransactionID) {
				return entriesSlice[i].Outpoint.Index < entriesSlice[j].Outpoint.Index
			}
			return entriesSlice[i].Outpoint.TransactionID.Less(&entriesSlice[j].Outpoint.TransactionID)
		}
		return entriesSlice[i].DAAScore < entriesSlice[j].DAAScore
	})
	// Find batch start and end points
	startIndex := len(entriesSlice)
	endIndex := uint64(len(entriesSlice))
	for i := 0; i < len(entriesSlice); i++ {
		if entriesSlice[i].DAAScore >= batchDaaScoreStart {
			startIndex = i
			break
		}
	}
	if uint64(startIndex)+batchSize < endIndex {
		endIndex = uint64(startIndex) + batchSize
	}
	return entriesSlice[startIndex:endIndex]
}
