package rpchandlers

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/mempool"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

// HandleSendRawTransaction handles the respectively named RPC command
func HandleSendRawTransaction(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	sendRawTransactionRequest := request.(*appmessage.SendRawTransactionRequestMessage)

	serializedTx, err := hex.DecodeString(sendRawTransactionRequest.TransactionHex)
	if err != nil {
		errorMessage := &appmessage.SendRawTransactionResponseMessage{}
		errorMessage.Error = &appmessage.RPCError{
			Message: fmt.Sprintf("Transaction hex could not be parsed: %s", err),
		}
		return errorMessage, nil
	}
	var msgTx appmessage.MsgTx
	err = msgTx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		errorMessage := &appmessage.SendRawTransactionResponseMessage{}
		errorMessage.Error = &appmessage.RPCError{
			Message: fmt.Sprintf("Transaction decode failed: %s", err),
		}
		return errorMessage, nil
	}

	tx := util.NewTx(&msgTx)
	err = context.ProtocolManager.AddTransaction(tx)
	if err != nil {
		if !errors.As(err, &mempool.RuleError{}) {
			return nil, err
		}

		log.Debugf("Rejected transaction %s: %s", tx.ID(), err)
		errorMessage := &appmessage.SendRawTransactionResponseMessage{}
		errorMessage.Error = &appmessage.RPCError{
			Message: fmt.Sprintf("Rejected transaction %s: %s", tx.ID(), err),
		}
		return errorMessage, nil
	}

	response := appmessage.NewSendRawTransactionResponseMessage(tx.ID().String())
	return response, nil
}
