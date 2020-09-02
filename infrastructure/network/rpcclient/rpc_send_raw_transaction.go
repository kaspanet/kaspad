package rpcclient

import (
	"bytes"
	"encoding/hex"
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (c *RPCClient) SendRawTransaction(msgTx *appmessage.MsgTx) (*appmessage.SendRawTransactionResponseMessage, error) {
	transactionHex := ""
	if msgTx != nil {
		buf := bytes.NewBuffer(make([]byte, 0, msgTx.SerializeSize()))
		if err := msgTx.Serialize(buf); err != nil {
			return nil, err
		}
		transactionHex = hex.EncodeToString(buf.Bytes())
	}
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewSendRawTransactionRequestMessage(transactionHex))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdSendRawTransactionResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	sendRawTransactionResponse := response.(*appmessage.SendRawTransactionResponseMessage)
	if sendRawTransactionResponse.Error != nil {
		return nil, c.convertRPCError(sendRawTransactionResponse.Error)
	}

	return sendRawTransactionResponse, nil
}
