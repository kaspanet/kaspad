package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util"
)

// HandleGetMempoolEntriesByAddresses handles the respectively named RPC command
func HandleGetMempoolEntriesByAddresses(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {

	getMempoolEntriesByAddressesRequest := request.(*appmessage.GetMempoolEntriesByAddressesRequestMessage)

	mempoolEntriesByAddresses := make([]*appmessage.MempoolEntryByAddress, 0)

	for _, addressString := range getMempoolEntriesByAddressesRequest.Addresses {

		address, err := util.DecodeAddress(addressString, context.Config.NetParams().Prefix)
		if err != nil {
			errorMessage := &appmessage.GetMempoolEntriesByAddressesResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("Could not decode address '%s': %s", addressString, err)
			return errorMessage, nil
		}

		sending := make([]*appmessage.MempoolEntry, 0)
		receiving := make([]*appmessage.MempoolEntry, 0)

		if !getMempoolEntriesByAddressesRequest.FilterTransactionPool {

			sendingInMempool, receivingInMempool, err := context.Domain.MiningManager().GetTransactionsByAddresses()
			if err != nil {
				return nil, err
			}

			if transaction, found := sendingInMempool[address]; found {
				rpcTransaction := appmessage.DomainTransactionToRPCTransaction(transaction)
				sending = append(sending, &appmessage.MempoolEntry{
					Fee:         transaction.Fee,
					Transaction: rpcTransaction,
					IsOrphan:    false})
			}

			if transaction, found := receivingInMempool[address]; found {
				rpcTransaction := appmessage.DomainTransactionToRPCTransaction(transaction)
				receiving = append(receiving, &appmessage.MempoolEntry{
					Fee:         transaction.Fee,
					Transaction: rpcTransaction,
					IsOrphan:    false})
			}
		}
		if getMempoolEntriesByAddressesRequest.IncludeOrphanPool {

			sendingInOrphanPool, receivingInOrphanPool, err := context.Domain.MiningManager().GetOrphanTransactionsByAddresses()
			if err != nil {
				return nil, err
			}

			if transaction, found := sendingInOrphanPool[address]; found {
				rpcTransaction := appmessage.DomainTransactionToRPCTransaction(transaction)
				sending = append(sending, &appmessage.MempoolEntry{
					Fee:         transaction.Fee,
					Transaction: rpcTransaction,
					IsOrphan:    true})
			}

			if transaction, found := receivingInOrphanPool[address]; found {
				rpcTransaction := appmessage.DomainTransactionToRPCTransaction(transaction)
				receiving = append(receiving, &appmessage.MempoolEntry{
					Fee:         transaction.Fee,
					Transaction: rpcTransaction,
					IsOrphan:    true})
			}

		}

		if len(sending) > 0 || len(receiving) > 0 {
			mempoolEntriesByAddresses = append(
				mempoolEntriesByAddresses,
				&appmessage.MempoolEntryByAddress{
					Address:   address.String(),
					Sending:   sending,
					Receiving: receiving,
				},
			)
		}
	}

	return appmessage.NewGetMempoolEntriesByAddressesResponseMessage(mempoolEntriesByAddresses), nil
}
