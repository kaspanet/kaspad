package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/addressindex"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetTransactionsByAddresses handles the respectively named RPC command
func HandleGetTransactionsByAddresses(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getTransactionsRequest := request.(*appmessage.GetTransactionsByAddressesRequestMessage)
	startingBlockHash := getTransactionsRequest.StartingBlockHash
	addressMap := make(map[string]struct{})
	for _, address := range getTransactionsRequest.Addresses {
		addressMap[address] = struct{}{}
	}

	var lowHash *externalapi.DomainHash
	if len(startingBlockHash) > 0 {
		hash, err := hashes.FromString(getTransactionsRequest.StartingBlockHash)
		if err != nil {
			errorMessage := &appmessage.GetTransactionsByAddressesResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("Hash could not be parsed: %s", err)
			return errorMessage, nil
		}
		lowHash = hash
	} else {
		lowHash = context.Config.NetParams().GenesisHash
	}

	_, err := context.Domain.Consensus().GetBlock(lowHash)
	if err != nil {
		errorMessage := &appmessage.GetBlockResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Block %s not found", lowHash)
		return errorMessage, nil
	}

	blockHashes, err := context.Domain.Consensus().GetHashesBetween(lowHash, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	if len(blockHashes) == 0 {
		return appmessage.NewGetTransactionsByAddressesResponseMessage(startingBlockHash, nil), nil
	}

	lastBlockHash := blockHashes[len(blockHashes)-1]

	var transactionsVerboseData []*appmessage.TransactionVerboseData
	for _, blockHash := range blockHashes {
		block, err := context.Domain.Consensus().GetBlock(blockHash)
		if err != nil {
			return nil, err
		}
		for _, transaction := range block.Transactions {
			for _, txOut := range transaction.Outputs {
				address, err := addressindex.GetAddress(txOut.ScriptPublicKey, context.Config.NetParams().Prefix)
				if err != nil {
					return nil, err
				}
				if _, ok := addressMap[address]; ok {
					txID := consensusserialization.TransactionID(transaction).String()
					transactionVerboseData, err := context.BuildTransactionVerboseData(transaction, txID, block.Header, blockHash.String())
					if err != nil {
						return nil, err
					}
					transactionsVerboseData = append(transactionsVerboseData, transactionVerboseData)
					break
				}
			}
		}
	}

	response := appmessage.NewGetTransactionsByAddressesResponseMessage(lastBlockHash.String(), transactionsVerboseData)
	return response, nil
}
