package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util"
)

// HandleGetMempoolEntriesByAddresses handles the respectively named RPC command
func HandleGetMempoolEntriesByAddresses(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getMempoolEntriesByAddressesRequest := request.(*appmessage.GetMempoolEntriesByAddressesRequestMessage)
	transactions := context.Domain.MiningManager().AllTransactions()

	addressStrings := getMempoolEntriesByAddressesRequest.Addresses
	scriptPublicKeys := make([]*externalapi.ScriptPublicKey, len(addressStrings))
	for i, addressString := range addressStrings {
		address, err := util.DecodeAddress(addressString, context.Config.NetParams().Prefix)
		if err != nil {
			errorMessage := &appmessage.GetUTXOsByAddressesResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("Could not decode address '%s': %s", addressString, err)
			return errorMessage, nil
		}
		scriptPublicKeys[i], err = txscript.PayToAddrScript(address)
		if err != nil {
			errorMessage := &appmessage.GetUTXOsByAddressesResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("Could not create a scriptPublicKey for address '%s': %s", addressString, err)
			return errorMessage, nil
		}
	}

	spendingEntries := []*appmessage.MempoolEntry{}
	receivingEntries := []*appmessage.MempoolEntry{}
	for _, transaction := range transactions {
		isSpending := isSpendingTransaction(transaction, scriptPublicKeys)
		isReceiving := isReceivingTransaction(transaction, scriptPublicKeys)

		if !isSpending && !isReceiving {
			continue
		}

		rpcTransaction := appmessage.DomainTransactionToRPCTransaction(transaction)
		err := context.PopulateTransactionWithVerboseData(rpcTransaction, nil)
		if err != nil {
			return nil, err
		}

		if isSpending {
			spendingEntries = append(spendingEntries, &appmessage.MempoolEntry{
				Fee:         transaction.Fee,
				Transaction: rpcTransaction,
			})
		}
		if isReceiving {
			receivingEntries = append(receivingEntries, &appmessage.MempoolEntry{
				Fee:         transaction.Fee,
				Transaction: rpcTransaction,
			})
		}
	}

	return appmessage.NewGetMempoolEntriesByAddressesResponseMessage(spendingEntries, receivingEntries), nil
}

func isSpendingTransaction(transaction *externalapi.DomainTransaction,
	scriptPublicKeys []*externalapi.ScriptPublicKey) bool {

	for _, input := range transaction.Inputs {
		for _, scriptPublicKey := range scriptPublicKeys {
			if scriptPublicKey == input.UTXOEntry.ScriptPublicKey() {
				return true
			}
		}
	}
	return false
}

func isReceivingTransaction(transaction *externalapi.DomainTransaction,
	scriptPublicKeys []*externalapi.ScriptPublicKey) bool {

	for _, input := range transaction.Inputs {
		for _, scriptPublicKey := range scriptPublicKeys {
			if scriptPublicKey == input.UTXOEntry.ScriptPublicKey() {
				return true
			}
		}
	}
	return false
}
