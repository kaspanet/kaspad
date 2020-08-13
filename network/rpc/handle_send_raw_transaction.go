package rpc

import (
	"bytes"
	"encoding/hex"
	"github.com/kaspanet/kaspad/domain/mempool"
	"github.com/kaspanet/kaspad/network/domainmessage"
	"github.com/kaspanet/kaspad/network/rpc/model"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

// handleSendRawTransaction implements the sendRawTransaction command.
func handleSendRawTransaction(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*model.SendRawTransactionCmd)
	// Deserialize and send off to tx relay
	hexStr := c.HexTx
	serializedTx, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, rpcDecodeHexError(hexStr)
	}
	var msgTx domainmessage.MsgTx
	err = msgTx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		return nil, &model.RPCError{
			Code:    model.ErrRPCDeserialization,
			Message: "TX decode failed: " + err.Error(),
		}
	}

	tx := util.NewTx(&msgTx)
	err = s.protocolManager.AddTransaction(tx)
	if err != nil {
		if !errors.As(err, &mempool.RuleError{}) {
			panic(err)
		}

		log.Debugf("Rejected transaction %s: %s", tx.ID(), err)
		return nil, &model.RPCError{
			Code:    model.ErrRPCVerify,
			Message: "TX rejected: " + err.Error(),
		}
	}

	return tx.ID().String(), nil
}
