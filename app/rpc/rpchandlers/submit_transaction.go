package rpchandlers

import (
	"bytes"
	"encoding/hex"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"
	"github.com/kaspanet/kaspad/domain/mempool"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// HandleSubmitTransaction handles the respectively named RPC command
func HandleSubmitTransaction(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	submitTransactionRequest := request.(*appmessage.SubmitTransactionRequestMessage)

	serializedTx, err := hex.DecodeString(submitTransactionRequest.TransactionHex)
	if err != nil {
		errorMessage := &appmessage.SubmitTransactionResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Transaction hex could not be parsed: %s", err)
		return errorMessage, nil
	}

	msgTx := &appmessage.MsgTx{}
	err = msgTx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		errorMessage := &appmessage.SubmitTransactionResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Transaction decode failed: %s", err)
		return errorMessage, nil
	}

	domainTransaction := appmessage.MsgTxToDomainTransaction(msgTx)
	transactionID := hashserialization.TransactionID(domainTransaction)
	err = context.ProtocolManager.AddTransaction(domainTransaction)
	if err != nil {
		if !errors.As(err, &mempool.RuleError{}) {
			return nil, err
		}

		log.Debugf("Rejected transaction %s: %s", transactionID, err)
		errorMessage := &appmessage.SubmitTransactionResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Rejected transaction %s: %s", transactionID, err)
		return errorMessage, nil
	}

	response := appmessage.NewSubmitTransactionResponseMessage(transactionID.String())
	return response, nil
}
