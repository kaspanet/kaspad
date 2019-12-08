package rpc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/btcjson"
	"github.com/kaspanet/kaspad/mempool"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

// handleSendRawTransaction implements the sendRawTransaction command.
func handleSendRawTransaction(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.SendRawTransactionCmd)
	// Deserialize and send off to tx relay
	hexStr := c.HexTx
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	serializedTx, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, rpcDecodeHexError(hexStr)
	}
	var msgTx wire.MsgTx
	err = msgTx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCDeserialization,
			Message: "TX decode failed: " + err.Error(),
		}
	}

	// Use 0 for the tag to represent local node.
	tx := util.NewTx(&msgTx)
	acceptedTxs, err := s.cfg.TxMemPool.ProcessTransaction(tx, false, 0)
	if err != nil {
		// When the error is a rule error, it means the transaction was
		// simply rejected as opposed to something actually going wrong,
		// so log it as such.  Otherwise, something really did go wrong,
		// so log it as an actual error.  In both cases, a JSON-RPC
		// error is returned to the client with the deserialization
		// error code (to match bitcoind behavior).
		if _, ok := err.(mempool.RuleError); ok {
			log.Debugf("Rejected transaction %s: %s", tx.ID(),
				err)
		} else {
			log.Errorf("Failed to process transaction %s: %s",
				tx.ID(), err)
		}
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCVerify,
			Message: "TX rejected: " + err.Error(),
		}
	}

	// When the transaction was accepted it should be the first item in the
	// returned array of accepted transactions.  The only way this will not
	// be true is if the API for ProcessTransaction changes and this code is
	// not properly updated, but ensure the condition holds as a safeguard.
	//
	// Also, since an error is being returned to the caller, ensure the
	// transaction is removed from the memory pool.
	if len(acceptedTxs) == 0 || !acceptedTxs[0].Tx.ID().IsEqual(tx.ID()) {
		err := s.cfg.TxMemPool.RemoveTransaction(tx, true, true)
		if err != nil {
			return nil, err
		}

		errStr := fmt.Sprintf("transaction %s is not in accepted list",
			tx.ID())
		return nil, internalRPCError(errStr, "")
	}

	// Generate and relay inventory vectors for all newly accepted
	// transactions into the memory pool due to the original being
	// accepted.
	s.cfg.ConnMgr.RelayTransactions(acceptedTxs)

	// Notify both websocket and getBlockTemplate long poll clients of all
	// newly accepted transactions.
	s.NotifyNewTransactions(acceptedTxs)

	// Keep track of all the sendRawTransaction request txns so that they
	// can be rebroadcast if they don't make their way into a block.
	txD := acceptedTxs[0]
	iv := wire.NewInvVect(wire.InvTypeTx, (*daghash.Hash)(txD.Tx.ID()))
	s.cfg.ConnMgr.AddRebroadcastInventory(iv, txD)

	return tx.ID().String(), nil
}
