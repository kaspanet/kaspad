package rpc

import (
	"bytes"
	"encoding/hex"
	"github.com/kaspanet/kaspad/kaspajson"
	"github.com/kaspanet/kaspad/wire"
)

// handleDecodeRawTransaction handles decodeRawTransaction commands.
func handleDecodeRawTransaction(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*kaspajson.DecodeRawTransactionCmd)

	// Deserialize the transaction.
	hexStr := c.HexTx
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	serializedTx, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, rpcDecodeHexError(hexStr)
	}
	var mtx wire.MsgTx
	err = mtx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		return nil, &kaspajson.RPCError{
			Code:    kaspajson.ErrRPCDeserialization,
			Message: "TX decode failed: " + err.Error(),
		}
	}

	// Create and return the result.
	txReply := kaspajson.TxRawDecodeResult{
		TxID:     mtx.TxID().String(),
		Version:  mtx.Version,
		Locktime: mtx.LockTime,
		Vin:      createVinList(&mtx),
		Vout:     createVoutList(&mtx, s.cfg.DAGParams, nil),
	}
	return txReply, nil
}
