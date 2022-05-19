package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util"
)

// HandleGetMempoolEntriesByAddresses handles the respectively named RPC command
func HandleGetMempoolEntriesByAddresses(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {

	transactions := context.Domain.MiningManager().AllTransactions()
	getMempoolEntriesByAddressesRequest := request.(*appmessage.GetMempoolEntriesByAddressesRequestMessage)
	mempoolEntriesByAddresses := make([]*appmessage.MempoolEntryByAddress, 0)

	for _, addressString := range getMempoolEntriesByAddressesRequest.Addresses {

		_, err := util.DecodeAddress(addressString, context.Config.ActiveNetParams.Prefix)
		if err != nil {
			errorMessage := &appmessage.GetUTXOsByAddressesResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("Could not decode address '%s': %s", addressString, err)
			return errorMessage, nil
		}

		sending := make([]*appmessage.MempoolEntry, 0)
		receiving := make([]*appmessage.MempoolEntry, 0)

		for _, transaction := range transactions {

			for i, input := range transaction.Inputs {
				// TODO: Fix this
				if input.UTXOEntry == nil {
					log.Errorf("Couldn't find UTXO entry for input %d in mempool transaction %s. This should never happen.", i, consensushashing.TransactionID(transaction))
					continue
				}

				_, transactionSendingAddress, err := txscript.ExtractScriptPubKeyAddress(
					input.UTXOEntry.ScriptPublicKey(),
					context.Config.ActiveNetParams)
				if err != nil {
					return nil, err
				}
				if addressString == transactionSendingAddress.String() {
					rpcTransaction := appmessage.DomainTransactionToRPCTransaction(transaction)
					sending = append(
						sending,
						&appmessage.MempoolEntry{
							Fee:         transaction.Fee,
							Transaction: rpcTransaction,
						},
					)
					break //one input is enough
				}
			}

			for _, output := range transaction.Outputs {
				_, transactionReceivingAddress, err := txscript.ExtractScriptPubKeyAddress(
					output.ScriptPublicKey,
					context.Config.ActiveNetParams,
				)
				if err != nil {
					return nil, err
				}
				if addressString == transactionReceivingAddress.String() {
					rpcTransaction := appmessage.DomainTransactionToRPCTransaction(transaction)
					receiving = append(
						receiving,
						&appmessage.MempoolEntry{
							Fee:         transaction.Fee,
							Transaction: rpcTransaction,
						},
					)
					break //one output is enough
				}
			}

			//Only append mempoolEntriesByAddress, if at least 1 mempoolEntry for the address is found.
			//This mimics the behaviour of GetUtxosByAddresses RPC call.
			if len(sending) > 0 || len(receiving) > 0 {
				mempoolEntriesByAddresses = append(
					mempoolEntriesByAddresses,
					&appmessage.MempoolEntryByAddress{
						Address:   addressString,
						Sending:   sending,
						Receiving: receiving,
					},
				)
			}

		}
	}

	return appmessage.NewGetMempoolEntriesByAddressesResponseMessage(mempoolEntriesByAddresses), nil
}
